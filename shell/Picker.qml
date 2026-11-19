import QtQuick
import Quickshell
import Quickshell.Hyprland
import Quickshell.Io
import Quickshell.Wayland
import qs.common

Scope {
  id: root

  readonly property int wide: 660
  readonly property int narrow: 96
  readonly property int gap: 12
  property string mode: ""
  property var items: []
  property string filter: ""
  property string selected: ""
  property int index: 0
  property bool busy: false
  property string trying: ""
  property string failed: ""
  property string why: ""
  property var said: []
  property string action: ""
  property int opening: 0
  property int startedIn: 0
  readonly property bool open: mode !== ""
  readonly property string yours: (Quickshell.env("XDG_CONFIG_HOME") || Quickshell.env("HOME") + "/.config") + "/solanum/backgrounds/"
  readonly property var current: shown[index] || null
  readonly property string hint: failed ? why || "Couldn't apply " + failed : said.length ? "Applied, with what was fixed below" : filter ? "Filter: " + filter : ""
  readonly property var hints: [[["Left", "Right"], "move"], ["Enter", current && current.value === "keep" ? "keep it in the theme's own wallpapers" : current && current.yours ? "show it" : "apply"]].concat(current && current.yours ? [["Delete", "to the Trash"]] : [], [["", "type to filter"], ["Esc", "close"]])
  readonly property var shown: items.filter(item => item.label.toLowerCase().includes(filter.toLowerCase()))
  readonly property var screen: Quickshell.screens.find(item => Hyprland.focusedMonitor && item.name === Hyprland.focusedMonitor.name) || Quickshell.screens[0]
  readonly property var drawn: [
    {
      value: "draw",
      label: "Draw another",
      image: "",
      glyph: "\uf1fc"
    },
    {
      value: "logo",
      label: "Animated logo",
      image: "",
      glyph: ""
    },
    {
      value: "life",
      label: "Game of Life",
      image: "",
      glyph: "\uf00a"
    },
    {
      value: "add",
      label: "Add pictures to this theme",
      image: "",
      glyph: "\uf067"
    },
    {
      value: "add-every",
      label: "Add pictures for every theme",
      image: "",
      glyph: "\uf067"
    }
  ]

  function show(kind) {
    mode = kind;
    filter = "";
    failed = "";
    said = [];
    busy = applying.running;
    opening++;
    items = [];
    content.forceActiveFocus();
    if (kind === "themes") {
      themeName.reload();
      selected = themeName.text().trim();
      themeList.running = true;
    } else {
      background.running = true;
    }
  }

  function take(kind, text) {
    if (mode !== kind)
      return;
    const drawing = kind === "backgrounds" && selected.includes("/.local/state/solanum/current/drawings/") ? [
      {
        value: "keep",
        label: "Keep this drawing",
        image: selected,
        glyph: ""
      }
    ] : [];
    items = drawing.concat(text.split("\n").filter(line => line !== "").map(line => parse(kind, line))).concat(kind === "backgrounds" ? drawn : []);
    index = Math.max(0, shown.findIndex(item => item.value === selected || item.value === "keep" || item.value === "logo" && Theme.looks.logo === "animated" || item.value === "life" && Theme.looks.logo === "life"));
  }

  function address(path) {
    return path ? "file://" + path.split("/").map(encodeURIComponent).join("/") : "";
  }

  function close() {
    mode = "";
  }

  function move(step) {
    failed = "";
    said = [];
    if (shown.length)
      index = Math.max(0, Math.min(shown.length - 1, index + step));
  }

  function key(event) {
    const steps = {
      [Qt.Key_Left]: -1,
      [Qt.Key_Up]: -1,
      [Qt.Key_Right]: 1,
      [Qt.Key_Down]: 1,
      [Qt.Key_PageUp]: -5,
      [Qt.Key_PageDown]: 5,
      [Qt.Key_Home]: -shown.length,
      [Qt.Key_End]: shown.length
    };
    if (steps[event.key] !== undefined)
      move(steps[event.key]);
    else if (event.key === Qt.Key_Escape)
      close();
    else if (event.key === Qt.Key_Return || event.key === Qt.Key_Enter)
      apply();
    else if (event.key === Qt.Key_Delete)
      remove();
    else if (event.key === Qt.Key_Backspace)
      filter = filter.slice(0, -1);
    else if (/^[\w .-]$/.test(event.text))
      filter += event.text;
    else
      return false;
    return true;
  }

  function run(item, command) {
    busy = true;
    failed = "";
    said = [];
    action = item.value;
    trying = item.label;
    startedIn = opening;
    applying.command = command;
    applying.running = true;
  }

  function apply() {
    const item = shown[index];
    if (!item || busy)
      return;
    if (item.value === "add" || item.value === "add-every") {
      close();
      adding.command = item.value === "add" ? ["solanum-background", "add"] : ["solanum-background", "add", "--every"];
      adding.running = true;
      return;
    }
    run(item, mode === "themes" ? ["solanum-theme-set", item.value] : ["solanum-background", item.value]);
  }

  function remove() {
    if (current && current.yours && !busy)
      run(Object.assign({}, current, {
        value: "remove"
      }), ["solanum-background", "remove", current.value]);
  }

  function parse(kind, line) {
    if (kind === "themes") {
      const fields = line.split("\t");
      return {
        value: fields[0],
        label: fields[1] || fields[0],
        image: fields[2] || ""
      };
    }
    return {
      value: line,
      label: line.split("/").pop().replace(/\.[^.]+$/, ""),
      image: line,
      yours: line.startsWith(yours)
    };
  }

  function settle(code) {
    const lines = (applyOut.text + "\n" + applyErr.text).split("\n").filter(line => line.trim() !== "");
    if (startedIn !== opening) {
      busy = false;
      return;
    }
    const again = mode === "backgrounds" && ["draw", "keep", "remove"].includes(action);
    if (code === 0 && again) {
      busy = false;
      show("backgrounds");
      return;
    }
    if (code === 0 && (lines.length === 0 || mode === "backgrounds")) {
      done.restart();
      return;
    }
    busy = false;
    if (code === 0) {
      said = lines;
      return;
    }
    failed = trying;
    why = lines.find(line => line.startsWith("Couldn't apply")) || (mode === "backgrounds" ? lines[lines.length - 1] || "" : "");
  }

  onFilterChanged: {
    index = 0;
    failed = "";
    said = [];
  }

  FileView {
    id: themeName

    path: Quickshell.env("HOME") + "/.local/state/solanum/current/theme.name"
    blockLoading: true
    printErrors: false
  }

  Process {
    id: background

    command: ["solanum-background", "--current"]
    stdout: StdioCollector {
      onStreamFinished: {
        if (root.mode !== "backgrounds")
          return;
        root.selected = text.trim();
        backgroundList.running = true;
      }
    }
  }

  Process {
    id: themeList

    command: ["solanum-theme-set", "--previews"]
    stdout: StdioCollector {
      onStreamFinished: root.take("themes", text)
    }
  }

  Process {
    id: backgroundList

    command: ["solanum-background", "--list"]
    stdout: StdioCollector {
      onStreamFinished: root.take("backgrounds", text)
    }
  }

  Process {
    id: applying

    property int code: 0
    property int waiting: 0

    function finished() {
      waiting--;
      if (waiting === 0)
        root.settle(code);
    }

    onStarted: waiting = 3
    onExited: exitCode => {
      code = exitCode;
      finished();
    }
    stdout: StdioCollector {
      id: applyOut

      onStreamFinished: applying.finished()
    }
    stderr: StdioCollector {
      id: applyErr

      onStreamFinished: applying.finished()
    }
  }

  Process {
    id: adding

    onExited: root.show("backgrounds")
  }

  Timer {
    id: done

    interval: 300
    onTriggered: {
      if (root.startedIn === root.opening)
        root.close();
    }
  }

  IpcHandler {
    target: "picker"

    function themes(): void {
      root.show("themes");
    }

    function backgrounds(): void {
      root.show("backgrounds");
    }

    function close(): void {
      root.close();
    }
  }

  PanelWindow {
    screen: root.screen
    visible: root.open || content.opacity > 0
    color: "transparent"
    exclusionMode: ExclusionMode.Ignore
    WlrLayershell.layer: WlrLayer.Overlay
    WlrLayershell.namespace: "solanum-picker"
    onVisibleChanged: {
      if (!visible)
        root.items = [];
    }
    WlrLayershell.keyboardFocus: root.open ? WlrKeyboardFocus.Exclusive : WlrKeyboardFocus.None

    anchors {
      top: true
      bottom: true
      left: true
      right: true
    }

    Rectangle {
      id: content

      anchors.fill: parent
      color: Qt.alpha(Theme.base, 0.85)
      opacity: root.open ? 1 : 0
      focus: true
      Keys.onPressed: event => event.accepted = root.key(event)

      Fade on opacity {}

      MouseArea {
        anchors.fill: parent
        onClicked: root.close()
        onWheel: wheel => {
          const steps = notches.steps(wheel);
          if (steps !== 0)
            root.move(-steps);
        }

        Notches {
          id: notches
        }
      }

      Item {
        id: carousel

        anchors.left: parent.left
        anchors.right: parent.right
        anchors.verticalCenter: parent.verticalCenter
        anchors.verticalCenterOffset: -30
        height: 420

        Repeater {
          model: root.shown

          Item {
            id: slice

            required property var modelData
            required property int index
            readonly property bool current: index === root.index
            readonly property int offset: index - root.index

            x: carousel.width / 2 + (current ? -root.wide / 2 : offset < 0 ? -root.wide / 2 + offset * (root.narrow + root.gap) : root.wide / 2 + root.gap + (offset - 1) * (root.narrow + root.gap))
            width: current ? root.wide : root.narrow
            height: carousel.height

            Behavior on x {
              NumberAnimation {
                duration: Theme.slide
                easing.type: Easing.OutCubic
              }
            }

            Behavior on width {
              NumberAnimation {
                duration: Theme.slide
                easing.type: Easing.OutCubic
              }
            }

            Rectangle {
              anchors.fill: parent
              radius: Theme.radius
              color: ["logo", "life"].includes(slice.modelData.value) ? Theme.base : Theme.surface
              border.width: slice.current ? 2 : 0
              border.color: Theme.accent
              clip: true

              Image {
                anchors.fill: parent
                anchors.margins: slice.current ? 2 : 0
                source: root.address(slice.modelData.image)
                fillMode: Image.PreserveAspectCrop
                asynchronous: true
                sourceSize.width: root.wide
                sourceSize.height: carousel.height
              }

              Label {
                anchors.centerIn: parent
                visible: !slice.modelData.image
                text: slice.modelData.glyph || ""
                color: Theme.subtext
                font.pixelSize: Theme.size(slice.current ? 64 : 24)
              }

              Loader {
                anchors.centerIn: parent
                active: slice.modelData.value === "logo"

                sourceComponent: LogoCanvas {
                  cell: Math.min(slice.width, slice.height) * 0.8 / Logo.cells
                  resting: !slice.current
                }
              }

              Loader {
                anchors.fill: parent
                anchors.margins: 2
                active: slice.modelData.value === "life" && slice.current
                clip: true

                sourceComponent: Item {
                  LifePlayer {
                    x: (parent.width - width * scale) / 2
                    y: (parent.height - height * scale) / 2
                    name: "picker"
                    cols: Math.ceil(root.wide / 6)
                    rows: Math.ceil(carousel.height / 6)
                    scale: 6
                  }
                }
              }

              Rectangle {
                anchors.fill: parent
                color: Theme.base
                opacity: slice.current ? 0 : 0.5

                Behavior on opacity {
                  NumberAnimation {
                    duration: Theme.slide
                  }
                }
              }
            }

            MouseArea {
              anchors.fill: parent
              cursorShape: Qt.PointingHandCursor
              onClicked: {
                if (slice.current)
                  root.apply();
                else
                  root.index = slice.index;
              }
            }
          }
        }
      }

      Label {
        id: title

        anchors.top: carousel.bottom
        anchors.topMargin: 24
        anchors.horizontalCenter: parent.horizontalCenter
        text: root.shown.length ? (root.busy ? "Applying " : "") + root.shown[root.index].label : root.items.length ? "Nothing matches " + root.filter : ""
        font.pixelSize: Theme.size(22)
        font.bold: true
      }

      Label {
        anchors.left: title.right
        anchors.leftMargin: 12
        anchors.baseline: title.baseline
        visible: root.shown.length > 0
        text: (root.index + 1) + " of " + root.shown.length
        color: Theme.subtext
      }

      Label {
        id: hintLine

        anchors.top: title.bottom
        anchors.topMargin: 10
        anchors.horizontalCenter: parent.horizontalCenter
        width: Math.min(implicitWidth, parent.width - 80)
        height: text ? implicitHeight : 0
        text: root.hint
        color: root.failed ? Theme.critical : Theme.subtext
        elide: Text.ElideRight
      }

      Hints {
        id: keyLine

        anchors.top: hintLine.bottom
        anchors.topMargin: 12
        anchors.horizontalCenter: parent.horizontalCenter
        pairs: root.hints
      }

      Label {
        id: trash

        anchors.top: keyLine.bottom
        anchors.topMargin: 10
        anchors.horizontalCenter: parent.horizontalCenter
        visible: !!root.current && !!root.current.yours && !root.busy && !root.failed
        text: "\uf1f8  Move to the Trash"
        color: trashArea.containsMouse ? Theme.critical : Theme.subtext

        MouseArea {
          id: trashArea

          anchors.fill: parent
          hoverEnabled: true
          cursorShape: Qt.PointingHandCursor
          onClicked: root.remove()
        }
      }

      Column {
        anchors.top: trash.visible ? trash.bottom : keyLine.bottom
        anchors.topMargin: 14
        anchors.horizontalCenter: parent.horizontalCenter
        width: Math.min(900, parent.width - 80)
        spacing: 6

        Repeater {
          model: root.said

          Label {
            required property string modelData

            width: parent.width
            text: modelData
            wrapMode: Text.Wrap
            horizontalAlignment: Text.AlignHCenter
          }
        }
      }
    }
  }
}
