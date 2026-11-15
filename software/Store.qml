pragma Singleton

import QtQuick
import Quickshell
import Quickshell.Io
import qs.common

Singleton {
  id: store

  readonly property var pages: [
    {
      id: "apps",
      label: "Apps",
      glyph: "",
      help: "SolanumOS's own picks, by kind. Enter installs or removes one, Space marks several."
    },
    {
      id: "packages",
      label: "Debian packages",
      glyph: "",
      help: "Every package Debian has, with its one-line description. I shows only the installed ones."
    },
    {
      id: "webapps",
      label: "Web apps",
      glyph: "",
      help: "Websites that open in a window of their own, from the launcher."
    },
    {
      id: "running",
      label: "Running",
      glyph: "",
      help: "What Software is installing or removing now, and how it went."
    }
  ]
  readonly property int most: 200
  readonly property int perJob: 100
  readonly property string helper: (Quickshell.env("SOLANUM_BIN_DIR") || "/usr/local/bin") + "/solanum-task-helper"
  property string page: pages.some(each => each.id === Quickshell.env("SOLANUM_SOFTWARE_PAGE")) ? Quickshell.env("SOLANUM_SOFTWARE_PAGE") : "apps"
  property string wanted: Quickshell.env("SOLANUM_SOFTWARE_ITEM") || ""
  property string search: ""
  property var apps: []
  property bool appsRead: false
  property var packages: []
  property bool packagesRead: false
  property bool installedOnly: false
  property var webapps: []
  property var marks: ({})
  property var queue: []
  property var job: null
  property string said: ""
  property bool failed: false
  readonly property var pageNow: pages.find(each => each.id === page) || pages[0]
  readonly property int marked: Object.keys(marks).length
  readonly property bool working: task.running || queue.length > 0
  readonly property Task task: Task {
    onRunningChanged: {
      if (!running)
        Qt.callLater(store.next);
    }
  }

  function show(id) {
    if (pages.some(each => each.id === id))
      page = id;
  }

  function parse(text, fallback) {
    try {
      return JSON.parse(text);
    } catch (error) {
      return fallback;
    }
  }

  function load() {
    appLister.running = true;
    webLister.running = true;
    packageLister.running = true;
  }

  function key(kind, id) {
    return kind + ":" + id;
  }

  function mark(kind, id, name, installed) {
    const next = Object.assign({}, marks);
    const at = key(kind, id);
    if (next[at])
      delete next[at];
    else
      next[at] = {
        kind: kind,
        id: id,
        name: name,
        install: !installed
      };
    marks = next;
  }

  function isMarked(kind, id) {
    return !!marks[key(kind, id)];
  }

  function words(text) {
    return text.toLowerCase().split(/\s+/).filter(word => word !== "");
  }

  function score(item, list) {
    let total = 0;
    for (const word of list) {
      const at = item.name.indexOf(word);
      if (at === 0)
        total += item.name.length === word.length ? 0 : 1;
      else if (at > 0)
        total += 2;
      else if (item.low.indexOf(word) >= 0)
        total += 3;
      else
        return -1;
    }
    return total;
  }

  function packageMatches(text) {
    const list = words(text);
    const pool = installedOnly ? packages.filter(item => item.installed) : packages;
    if (list.length === 0)
      return installedOnly ? pool : [];
    const buckets = Array.from({
      length: list.length * 3 + 1
    }, () => []);
    const count = pool.length;
    for (let at = 0; at < count; at++) {
      const item = pool[at];
      const total = score(item, list);
      if (total >= 0 && buckets[total].length < most)
        buckets[total].push(item);
    }
    return [].concat(...buckets).slice(0, most);
  }

  function appMatches(text) {
    const list = words(text);
    return apps.filter(app => list.every(word => (app.name + " " + app.id + " " + app.group).toLowerCase().includes(word)));
  }

  function chunked(list) {
    const found = [];
    for (let at = 0; at < list.length; at += perJob)
      found.push(list.slice(at, at + perJob));
    return found;
  }

  function changes() {
    const all = Object.values(marks);
    const apps = all.filter(each => each.kind === "app").map(each => ({
          job: each.install ? "app-install" : "app-remove",
          args: [each.id],
          label: (each.install ? "Install " : "Remove ") + each.name
        }));
    const packages = [].concat(...[true, false].map(install => chunked(all.filter(each => each.kind === "pkg" && each.install === install)))).map(group => ({
          job: group[0].install ? "pkg-install" : "pkg-remove",
          args: group.map(each => each.id),
          label: (group[0].install ? "Install " : "Remove ") + group.map(each => each.id).join(", ")
        }));
    return apps.concat(packages);
  }

  function apply(list) {
    if (list.length === 0)
      return;
    marks = {};
    page = "running";
    if (working) {
      queue = queue.concat(list);
      return;
    }
    queue = list;
    job = null;
    next();
  }

  function next() {
    if (task.running)
      return;
    const failed = job !== null && task.code !== 0 && task.code !== -1;
    if (queue.length === 0 || failed) {
      queue = [];
      load();
      return;
    }
    job = queue[0];
    queue = queue.slice(1);
    task.command = ["pkexec", helper, job.job].concat(job.args);
    task.start();
  }

  function addWebApp(name, address, icon) {
    said = "";
    failed = false;
    webChanger.command = ["solanum-webapp", "add", name, address].concat(icon ? [icon] : []);
    webChanger.running = true;
  }

  function removeWebApp(name) {
    said = "";
    failed = false;
    webChanger.command = ["solanum-webapp", "remove", name];
    webChanger.running = true;
  }

  Process {
    id: appLister

    command: ["solanum", "app", "list", "--json"]
    stdout: StdioCollector {
      onStreamFinished: {
        store.apps = store.parse(text, store.apps);
        store.appsRead = true;
      }
    }
  }

  Process {
    id: packageLister

    command: ["solanum-pkg", "search", "--json"]
    stdout: StdioCollector {
      onStreamFinished: {
        store.packages = store.parse(text, []).map(item => Object.assign(item, {
              low: item.description.toLowerCase()
            }));
        store.packagesRead = true;
      }
    }
  }

  Process {
    id: webLister

    command: ["solanum-webapp", "list", "--json"]
    stdout: StdioCollector {
      onStreamFinished: store.webapps = store.parse(text, store.webapps)
    }
  }

  Process {
    id: webChanger

    stdout: StdioCollector {
      onStreamFinished: {
        if (text.trim())
          store.said = text.trim();
      }
    }
    stderr: StdioCollector {
      onStreamFinished: {
        if (text.trim()) {
          store.said = text.trim().split("\n").pop();
          store.failed = true;
        }
      }
    }
    onExited: code => {
      store.failed = code !== 0;
      webLister.running = true;
    }
  }
}
