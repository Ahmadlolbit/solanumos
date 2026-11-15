pragma Singleton

import QtQuick
import Quickshell
import Quickshell.Io

Singleton {
  id: state

  property var pages: []
  property var rows: []
  property var about: ({
      name: "SolanumOS Chert",
      version: "",
      debian: "",
      facts: []
    })
  property var credits: []
  property var themes: []
  property string current: ""
  property var pictures: []
  property var palette: null
  property string page: Quickshell.env("SOLANUM_SETTINGS_PAGE") || "about"
  property string search: ""
  property string busy: ""
  property var said: []
  property string result: ""
  property string word: ""
  readonly property string yours: (Quickshell.env("XDG_CONFIG_HOME") || Quickshell.env("HOME") + "/.config") + "/solanum/backgrounds/"
  readonly property var pageNow: pages.find(item => item.id === page) || {
    id: page,
    label: page,
    glyph: ""
  }
  readonly property var shown: search ? found(search) : rows.filter(row => row.page === page && !row.hidden)
  readonly property string currentLabel: pretty(current)

  function pretty(name) {
    const listed = themes.find(item => item.name === name);
    return listed ? listed.label : name.split("-").map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(" ");
  }

  function words(row) {
    const owner = pages.find(item => item.id === row.page);
    return [row.label, row.help || "", (row.keywords || []).join(" "), owner ? owner.label + " " + (owner.keywords || []).join(" ") : ""].join(" ").toLowerCase();
  }

  function found(text) {
    const needle = text.toLowerCase().trim();
    return rows.filter(row => !row.hidden && words(row).includes(needle));
  }

  function parse(text, fallback) {
    try {
      return JSON.parse(text);
    } catch (error) {
      return fallback;
    }
  }

  function load() {
    lister.running = true;
    facts.running = true;
    makers.running = true;
    loadThemes();
  }

  function loadThemes() {
    themeList.running = true;
    shownTheme.reload();
    current = shownTheme.text().trim();
    pictureList.running = true;
  }

  function show(id) {
    search = "";
    page = id;
  }

  function set(row, value) {
    if (busy)
      return;
    busy = row.name;
    said = [];
    result = "";
    word = "";
    setter.command = ["solanum", "settings", "set", row.name, String(value)];
    setter.running = true;
  }

  function act(name, command, then) {
    if (busy)
      return;
    busy = name;
    said = [];
    result = "";
    word = "";
    actor.then = then || "";
    actor.command = command;
    actor.running = true;
  }

  function launch(command) {
    Quickshell.execDetached(command);
  }

  function heard(line) {
    if (line.trim() !== "")
      said = said.concat([line]);
  }

  function ended(code) {
    result = code === 0 ? "done" : "failed";
    busy = "";
    lister.running = true;
  }

  Component.onCompleted: load()

  FileView {
    id: shownTheme

    path: Quickshell.env("HOME") + "/.local/state/solanum/current/theme.name"
    blockLoading: true
    printErrors: false
    watchChanges: true
    onFileChanged: {
      reload();
      state.current = text().trim();
      pictureList.running = true;
    }
  }

  Process {
    id: lister

    command: ["solanum", "settings", "list"]
    stdout: StdioCollector {
      onStreamFinished: {
        const listing = state.parse(text, null);
        if (!listing)
          return;
        state.pages = listing.pages || [];
        state.rows = listing.rows || [];
      }
    }
  }

  Process {
    id: facts

    command: ["solanum", "settings", "about"]
    stdout: StdioCollector {
      onStreamFinished: state.about = state.parse(text, state.about)
    }
  }

  Process {
    id: makers

    command: ["solanum", "credits"]
    stdout: StdioCollector {
      onStreamFinished: state.credits = text.split("\n").filter(line => line.includes("\t")).map(line => line.split("\t"))
    }
  }

  Process {
    id: themeList

    command: ["solanum-theme-set", "--previews"]
    stdout: StdioCollector {
      onStreamFinished: state.themes = text.split("\n").filter(line => line.includes("\t")).map(line => {
        const fields = line.split("\t");
        return {
          name: fields[0],
          label: fields[1] || fields[0],
          image: fields[2] || ""
        };
      })
    }
  }

  Process {
    id: pictureList

    command: ["solanum-background", "--yours"]
    stdout: StdioCollector {
      onStreamFinished: state.pictures = text.split("\n").filter(line => line.startsWith(state.yours))
    }
  }

  Process {
    id: setter

    stdout: SplitParser {
      onRead: line => state.heard(line)
    }
    stderr: SplitParser {
      onRead: line => state.heard(line)
    }
    onExited: code => state.ended(code)
  }

  Process {
    id: actor

    property string then: ""

    stdout: SplitParser {
      onRead: line => state.heard(line)
    }
    stderr: SplitParser {
      onRead: line => state.heard(line)
    }
    onExited: code => {
      state.ended(code);
      if (then === "themes")
        state.loadThemes();
      else if (then === "pictures")
        pictureList.running = true;
    }
  }

  Process {
    id: paletteOf

    stdout: StdioCollector {
      onStreamFinished: state.palette = state.parse(text, null)
    }
  }

  function preview(command) {
    paletteOf.command = command;
    paletteOf.running = true;
  }
}
