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
  property int opening: 0
  property int startedIn: 0
  readonly property bool open: mode !== ""
  readonly property string hint: failed ? why || "Couldn't apply " + failed : said.length ? "Applied, with what was fixed below · Esc to close" : filter ? "Filter: " + filter : "Type to filter · arrows or wheel to move · Enter to apply · Esc to close"
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
    items = text.split("\n").filter(line => line !== "").map(line => parse(kind, line)).concat(kind === "backgrounds" ? drawn : []);
    index = Math.max(0, shown.findIndex(item => item.value === selected || item.value === "logo" && Theme.looks.logo === "animated"));
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

  function apply() {
    const item = shown[index];
    if (!item || busy)
      return;
    busy = true;
    failed = "";
    said = [];
    trying = item.label;
    startedIn = opening;
    applying.command = mode === "themes" ? ["solanum-theme-set", item.value] : ["solanum-background", item.value];
    applying.running = true;
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
      image: line
    };
  }

  function settle(code) {
    const lines = (applyOut.text + "\n" + applyErr.text).split("\n").filter(line => line.trim() !== "");
    if (startedIn !== opening) {
      busy = false;
      return;
    }
    if (code === 0 && lines.length === 0) {
      done.restart();
      return;
    }
    busy = false;
    if (code === 0) {
      said = lines;
      return;
    }
    failed = trying;
    why = lines.find(line => line.startsWith("Couldn't apply")) || "";
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
      Keys.onPressed: event => {
        if (event.key === Qt.Key_Escape)
          root.close();
        else if (event.key === Qt.Key_Left)
          root.move(-1);
        else if (event.key === Qt.Key_Right)
          root.move(1);
        else if (event.key === Qt.Key_Return || event.key === Qt.Key_Enter)
          root.apply();
        else if (event.key === Qt.Key_Backspace)
          root.filter = root.filter.slice(0, -1);
        else if (/^[\w .-]$/.test(event.text))
          root.filter += event.text;
        else
          return;
        event.accepted = true;
      }

      Behavior on opacity {
        NumberAnimation {
          duration: 250
          easing.type: Easing.OutCubic
        }
      }

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
                duration: 180
                easing.type: Easing.OutCubic
              }
            }

            Behavior on width {
              NumberAnimation {
                duration: 180
                easing.type: Easing.OutCubic
              }
            }

            Rectangle {
              anchors.fill: parent
              radius: Theme.radius
              color: slice.modelData.value === "logo" ? Theme.base : Theme.surface
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

              Rectangle {
                anchors.fill: parent
                color: Theme.base
                opacity: slice.current ? 0 : 0.5

                Behavior on opacity {
                  NumberAnimation {
                    duration: 180
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
        id: hintLine

        anchors.top: title.bottom
        anchors.topMargin: 10
        anchors.horizontalCenter: parent.horizontalCenter
        width: Math.min(implicitWidth, parent.width - 80)
        text: root.hint
        color: root.failed ? Theme.critical : Theme.subtext
        elide: Text.ElideRight
      }

      Column {
        anchors.top: hintLine.bottom
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
