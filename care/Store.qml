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
      label: "Updates",
      glyph: "\uf021",
      help: "What's waiting, and updating it with a snapshot taken first"
    },
    {
      id: "check",
      label: "Check",
      glyph: "\uf0f1",
      help: "The system check, and the safe fixes it offers"
    },
    {
      id: "changes",
      label: "What changed",
      glyph: "\uf1da",
      help: "What the last update of SolanumOS changed"
    },
    {
      id: "restore",
      label: "Restore points",
      glyph: "\uf0e2",
      help: "Going back to how the system was, with what changed since shown first"
    },
    {
      id: "space",
      label: "Space",
      glyph: "\uf0a0",
      help: "How full the disk is, and the old snapshots that can go"
    },
    {
      id: "undo",
      label: "Undo",
      glyph: "\uf112",
      help: "Taking back what a part of the setup installed"
    },
    {
      id: "report",
      label: "Report",
      glyph: "\uf188",
      help: "The logs and the system's details in one file, for an issue"
    }
  ]
  property string page: "updates"
  readonly property var pageNow: pages.find(item => item.id === page) || pages[0]
  property var waiting: null
  property var notes: null
  property var points: null
  property string pointsWhy: ""
  property var plan: null
  property var undos: null
  property string point: ""
  property var compared: ({})
  property string comparing: ""
  property string report: ""
  property string job: ""
  property string part: ""
  property bool holding: false
  readonly property string helper: Quickshell.env("SOLANUM_TASK_HELPER") || "/usr/local/bin/solanum-task-helper"
  property alias task: work
  readonly property bool running: work.running
  readonly property string stopWords: ({
      doctor: "Stop the check?",
      restore: "Stop going back?",
      cleanup: "Stop deleting snapshots?",
      undo: "Stop the undo?",
      collect: "Stop the report?"
    })[job] || "Stop the update?"

  function known(id) {
    return pages.some(item => item.id === String(id || "").split(":")[0]);
  }

  function show(id) {
    if (!known(id))
      return;
    const parts = String(id).split(":");
    page = parts[0];
    if (parts.length > 1)
      point = parts[1];
  }

  function load() {
    lister.running = true;
    reader.running = true;
    pointer.running = true;
    planner.running = true;
    undoer.running = true;
  }

  function compare(number) {
    if (!number || compared[number] || comparing)
      return;
    comparing = number;
    comparer.said = "";
    comparer.command = ["solanum", "restore", "compare", number, "--json"];
    comparer.running = true;
  }

  function settled() {
    const all = Object.assign({}, compared);
    all[comparing] = parsed(compareOut.text) || {
      failed: comparer.said || "The comparison stopped before it said anything"
    };
    compared = all;
    comparing = "";
    compare(point);
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

  Component.onCompleted: show(Quickshell.env("SOLANUM_CARE_PAGE"))

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
    id: pointer

    command: ["solanum", "restore", "list", "--json"]
    stdout: StdioCollector {
      onStreamFinished: store.points = store.parsed(text) || []
    }
    stderr: StdioCollector {
      onStreamFinished: store.pointsWhy = text.trim().split("\n").pop() || ""
    }
  }

  Process {
    id: planner

    command: ["solanum", "cleanup", "--plan", "--json"]
    stdout: StdioCollector {
      onStreamFinished: store.plan = store.parsed(text)
    }
  }

  Process {
    id: undoer

    command: ["solanum", "undo", "--list", "--json"]
    stdout: StdioCollector {
      onStreamFinished: store.undos = store.parsed(text) || []
    }
  }

  Process {
    id: comparer

    property string said: ""

    stdout: StdioCollector {
      id: compareOut
    }
    stderr: StdioCollector {
      onStreamFinished: comparer.said = text.trim().split("\n").pop() || ""
    }
    onRunningChanged: {
      if (!running)
        Qt.callLater(store.settled);
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
