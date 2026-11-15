pragma Singleton

import QtQuick
import Quickshell
import Quickshell.Io

Singleton {
  id: root

  property var history: []
  property var opened: []
  property var items: []
  property bool quiet: false
  property int seen: 0
  property int cleared: 0
  property var dismissed: []
  property real refreshed: 0
  readonly property var shownItems: items.filter(item => item.id > cleared && !dismissed.includes(item.id))
  readonly property int newest: items.length ? Math.max(...items.map(item => item.id)) : 0
  readonly property int unread: items.filter(item => item.id > Math.max(seen, cleared) && !dismissed.includes(item.id)).length
  readonly property string runtime: Quickshell.env("XDG_RUNTIME_DIR") || ""
  readonly property bool watching: watcher.running

  onSeenChanged: keeper.restart()
  onClearedChanged: keeper.restart()
  onDismissedChanged: keeper.restart()
  onNewestChanged: startedAgain()
  onItemsChanged: prune()
  Component.onCompleted: refresh()

  function character(match, name) {
    const named = {
      amp: "&",
      lt: "<",
      gt: ">",
      quot: "\"",
      apos: "'"
    };
    if (!name.startsWith("#"))
      return named[name];
    const code = name[1] === "x" ? parseInt(name.slice(2), 16) : parseInt(name.slice(1), 10);
    return code > 0 && code <= 0x10ffff ? String.fromCodePoint(code) : match;
  }

  function plain(body) {
    return body.replace(/<[^>]*>/g, "").replace(/&(amp|lt|gt|quot|apos|#[0-9]+|#x[0-9a-fA-F]+);/g, character);
  }

  function parse(text) {
    try {
      return JSON.parse(text).data[0].map(item => ({
            id: item.id.data,
            app: item["app-name"].data,
            summary: item.summary.data,
            body: plain(item.body.data),
            critical: item.urgency.data === 2
          })).reverse();
    } catch (error) {
      return [];
    }
  }

  function merged(current, past) {
    const ids = current.map(item => item.id);
    return current.concat(past.filter(item => !ids.includes(item.id))).sort((first, second) => second.id - first.id);
  }

  function startedAgain() {
    if (newest > 0 && newest < Math.max(seen, cleared)) {
      seen = 0;
      cleared = 0;
      dismissed = [];
    }
  }

  function restore(text) {
    try {
      const state = JSON.parse(text);
      seen = Math.max(seen, state.seen || 0);
      cleared = Math.max(cleared, state.cleared || 0);
      dismissed = dismissed.concat((state.dismissed || []).filter(id => !dismissed.includes(id)));
    } catch (error) {
      return;
    }
    startedAgain();
  }

  function refresh() {
    refreshed = Date.now();
    openLister.running = true;
    modes.running = true;
  }

  function open() {
    refresh();
    seen = newest;
  }

  function clear() {
    cleared = newest;
    dismissed = [];
  }

  function dismiss(id) {
    dismissed = dismissed.filter(known => known > cleared).concat([id]);
  }

  function prune() {
    const ids = items.map(item => item.id);
    if (dismissed.some(id => !ids.includes(id)))
      dismissed = dismissed.filter(id => ids.includes(id));
  }

  function heard(line) {
    if (String(line).includes("NameOwnerChanged"))
      dismissed = [];
    if (settle.running)
      return;
    settle.interval = Math.max(300, Math.min(2000, refreshed + 2000 - Date.now()));
    settle.start();
  }

  function toggleQuiet() {
    switcher.running = true;
  }

  Process {
    id: watcher

    running: true
    command: ["busctl", "--user", "monitor", "--json=short", "--match=type='method_call',interface='org.freedesktop.Notifications',member='Notify'", "--match=type='signal',interface='org.freedesktop.Notifications'", "--match=type='method_call',interface='fr.emersion.Mako',member='RestoreNotification'", "--match=type='method_call',interface='fr.emersion.Mako',member='SetModes'", "--match=type='method_call',interface='fr.emersion.Mako',member='SetMode'", "--match=type='signal',sender='org.freedesktop.DBus',member='NameOwnerChanged',arg0='org.freedesktop.Notifications'"]
    stdout: SplitParser {
      onRead: line => root.heard(line)
    }
  }

  Timer {
    id: settle

    onTriggered: root.refresh()
  }

  Process {
    id: openLister

    command: ["busctl", "--user", "-j", "call", "org.freedesktop.Notifications", "/fr/emersion/Mako", "fr.emersion.Mako", "ListNotifications"]
    stdout: StdioCollector {
      onStreamFinished: {
        root.opened = root.parse(text);
        lister.running = true;
      }
    }
  }

  Process {
    id: lister

    command: ["busctl", "--user", "-j", "call", "org.freedesktop.Notifications", "/fr/emersion/Mako", "fr.emersion.Mako", "ListHistory"]
    stdout: StdioCollector {
      onStreamFinished: {
        root.history = root.parse(text);
        root.items = root.merged(root.opened, root.history);
      }
    }
  }

  Process {
    id: modes

    command: ["makoctl", "mode"]
    stdout: StdioCollector {
      onStreamFinished: root.quiet = text.split("\n").includes("do-not-disturb")
    }
  }

  Process {
    id: switcher

    command: ["makoctl", "mode", "-t", "do-not-disturb"]
    onExited: modes.running = true
  }

  FileView {
    id: kept

    path: root.runtime ? root.runtime + "/solanum-notifications" : ""
    printErrors: false
    onLoaded: root.restore(text())
  }

  Timer {
    id: keeper

    interval: 100
    onTriggered: kept.setText(JSON.stringify({
        seen: root.seen,
        cleared: root.cleared,
        dismissed: root.dismissed
      }))
  }

  Timer {
    interval: 10000
    running: !root.watching
    repeat: true
    onTriggered: root.refresh()
  }
}
