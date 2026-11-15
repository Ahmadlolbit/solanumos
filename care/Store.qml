pragma Singleton

import QtQuick
import Quickshell
import Quickshell.Io
import qs.common

Singleton {
  id: store

  readonly property var pages: [
    {
      id: "updates",
      source: "Updates.qml",
      label: "Updates",
      glyph: "\uf021",
      help: "What's waiting, and updating it with a snapshot taken first"
    },
    {
      id: "check",
      source: "Check.qml",
      label: "Check",
      glyph: "\uf0f1",
      help: "The system check, and the safe fixes it offers"
    },
    {
      id: "changes",
      source: "Changes.qml",
      label: "What changed",
      glyph: "\uf1da",
      help: "What the last update of SolanumOS changed"
    }
  ]
  property string page: known(Quickshell.env("SOLANUM_CARE_PAGE")) ? Quickshell.env("SOLANUM_CARE_PAGE") : "updates"
  readonly property var pageNow: pages.find(item => item.id === page) || pages[0]
  property var waiting: null
  property var notes: null
  property string job: ""
  property string part: ""
  readonly property string helper: Quickshell.env("SOLANUM_TASK_HELPER") || "/usr/local/bin/solanum-task-helper"
  property alias task: work
  readonly property bool running: work.running

  function known(id) {
    return pages.some(item => item.id === id);
  }

  function show(id) {
    if (known(id))
      page = id;
  }

  function load() {
    lister.running = true;
    reader.running = true;
  }

  function parsed(text) {
    try {
      return JSON.parse(text);
    } catch (error) {
      return null;
    }
  }

  function start(name, args) {
    if (work.running)
      return;
    job = name;
    part = args.length > 1 ? args[1] : "";
    work.command = ["pkexec", helper].concat(args);
    work.start();
  }

  function finish() {
    job = "";
    part = "";
    load();
  }

  Task {
    id: work

    onRunningChanged: {
      if (running)
        return;
      store.load();
      if (store.job === "update")
        Quickshell.execDetached(["qs", "-p", "/usr/share/solanum/shell", "ipc", "call", "updates", "refresh"]);
    }
  }

  Process {
    id: lister

    command: ["solanum", "update", "--list", "--json"]
    stdout: StdioCollector {
      onStreamFinished: store.waiting = store.parsed(text) || {
        debian: [],
        apps: [],
        tools: [],
        solanum: [],
        firmware: [],
        last: "",
        restart: false,
        busy: false
      }
    }
  }

  Process {
    id: reader

    command: ["solanum", "changes", "--json"]
    stdout: StdioCollector {
      onStreamFinished: store.notes = store.parsed(text) || {
        date: "",
        groups: []
      }
    }
  }
}
