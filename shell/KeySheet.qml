import QtQuick
import QtQuick.Layouts
import Quickshell
import Quickshell.Hyprland
import Quickshell.Io
import Quickshell.Wayland
import qs.common

Scope {
  id: root

  property bool open: false
  property bool loaded: false
  property var binds: []
  property string filter: ""
  readonly property int rowHeight: Theme.size(28)
  readonly property var shown: binds.filter(item => (item.keys + " " + item.text).toLowerCase().includes(filter.toLowerCase()))
  readonly property var order: ["Apps", "Windows", "Workspaces", "Look", "Capture", "Media and hardware", "Yours"]
  readonly property var groups: [["Capture", /screenshot|record/], ["Media and hardware", /wpctl|playerctl|brightnessctl|volume|mute|brightness|track/], ["Look", /picker|background|solanum-look|theme|reader/], ["Apps", /kitty|xdg-open|gtk-launch|walker|hyprlock|ipc call (menu|keys)|solanum settings|terminal|files|browser|launcher|menu|settings|lock/], ["Workspaces", /workspace|scratchpad|special/], ["Windows", /window|focus|full ?screen|float|pseudo|split|resize|cycle|killactive/]]
  readonly property var sections: order.map(name => ({
        name: name,
        binds: shown.filter(item => item.group === name)
      })).filter(section => section.binds.length > 0)
  readonly property var screen: Quickshell.screens.find(item => Hyprland.focusedMonitor && item.name === Hyprland.focusedMonitor.name) || Quickshell.screens[0]
  readonly property var names: ({
      "return": "Enter",
      "escape": "Esc",
      "space": "Space",
      "tab": "Tab",
      "print": "Print",
      "mouse_down": "Wheel down",
      "mouse_up": "Wheel up",
      "xf86audioraisevolume": "Volume up",
      "xf86audiolowervolume": "Volume down",
      "xf86audiomute": "Mute",
      "xf86audiomicmute": "Mic mute",
      "xf86monbrightnessup": "Brightness up",
      "xf86monbrightnessdown": "Brightness down",
      "xf86audioplay": "Play",
      "xf86audiopause": "Pause",
      "xf86audionext": "Next track",
      "xf86audioprev": "Previous track",
      "left": "Left",
      "right": "Right",
      "up": "Up",
      "down": "Down"
    })
  readonly property var codes: ({
      "49": "the key left of 1"
    })

  function show() {
    filter = "";
    loaded = false;
    open = true;
    reader.running = true;
    list.contentY = 0;
    content.forceActiveFocus();
  }

  function close() {
    open = false;
  }

  function codeName(code) {
    const digits = ["1", "2", "3", "4", "5", "6", "7", "8", "9", "0"];
    if (code >= 10 && code <= 19)
      return digits[code - 10];
    return codes[code] || "Key " + code;
  }

  function keyName(bind) {
    const key = bind.key || "";
    if (!key && bind.keycode)
      return codeName(bind.keycode);
    if (key.startsWith("code:"))
      return codeName(parseInt(key.slice(5)));
    return names[key.toLowerCase()] || (key.length === 1 ? key.toUpperCase() : key);
  }

  function keysOf(bind) {
    const parts = [];
    if (bind.modmask & 64)
      parts.push("Super");
    if (bind.modmask & 4)
      parts.push("Ctrl");
    if (bind.modmask & 8)
      parts.push("Alt");
    if (bind.modmask & 1)
      parts.push("Shift");
    parts.push(keyName(bind));
    return parts.join(" + ");
  }

  function groupOf(bind) {
    const said = [bind.dispatcher, bind.arg, bind.description].join(" ").toLowerCase();
    const found = groups.find(group => group[1].test(said));
    return found ? found[0] : "Yours";
  }

  function parse(text) {
    let list = [];
    try {
      list = JSON.parse(text);
    } catch (error) {
      return [];
    }
    return list.filter(bind => bind.has_description && bind.description && !bind.mouse && !bind.submap).map(bind => ({
          keys: keysOf(bind),
          text: bind.description,
          group: groupOf(bind)
        })).filter((item, index, all) => all.findIndex(other => other.keys === item.keys) === index);
  }

  function scroll(by) {
    list.contentY = Math.max(0, Math.min(list.contentHeight - list.height, list.contentY + by));
  }

  function key(event) {
    const page = list.height - rowHeight;
    const steps = {
      [Qt.Key_Up]: -rowHeight,
      [Qt.Key_Down]: rowHeight,
      [Qt.Key_PageUp]: -page,
      [Qt.Key_PageDown]: page,
      [Qt.Key_Home]: -list.contentHeight,
      [Qt.Key_End]: list.contentHeight
    };
    if (steps[event.key] !== undefined)
      scroll(steps[event.key]);
    else if (event.key === Qt.Key_Escape)
      close();
    else if (event.key === Qt.Key_Backspace)
      filter = filter.slice(0, -1);
    else if (/^[\w .+-]$/.test(event.text))
      filter += event.text;
    else
      return false;
    return true;
  }

  onFilterChanged: list.contentY = 0

  Process {
    id: reader

    command: ["hyprctl", "binds", "-j"]
    stdout: StdioCollector {
      onStreamFinished: {
        root.binds = root.parse(text);
        root.loaded = true;
      }
    }
  }

  IpcHandler {
    target: "keys"

    function open(): void {
      root.show();
    }

    function close(): void {
      root.close();
    }

    function toggle(): void {
      if (root.open)
        root.close();
      else
        root.show();
    }
  }

  PanelWindow {
    screen: root.screen
    visible: root.open || content.opacity > 0
    color: "transparent"
    exclusionMode: ExclusionMode.Ignore
    WlrLayershell.layer: WlrLayer.Overlay
    WlrLayershell.namespace: "solanum-keys"
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

      Behavior on opacity {
        NumberAnimation {
          duration: 200
          easing.type: Easing.OutCubic
        }
      }

      MouseArea {
        anchors.fill: parent
        onClicked: root.close()
      }

      Rectangle {
        id: panel

        anchors.centerIn: parent
        width: Math.min(parent.width - 80, 1180)
        height: Math.min(parent.height - 120, header.implicitHeight + list.contentHeight + hints.implicitHeight + 84)
        color: Theme.panel
        border.color: Theme.accent
        border.width: 2
        radius: Theme.radius

        MouseArea {
          anchors.fill: parent
        }

        ColumnLayout {
          anchors.fill: parent
          anchors.margins: 24
          spacing: 18

          RowLayout {
            id: header

            Layout.fillWidth: true

            Label {
              text: "Keys"
              font.pixelSize: Theme.size(20)
              font.bold: true
            }

            Item {
              Layout.fillWidth: true
            }

            Label {
              visible: root.filter !== ""
              text: "Filter: " + root.filter
            }
          }

          Flickable {
            id: list

            Layout.fillWidth: true
            Layout.fillHeight: true
            contentHeight: Math.max(sheet.implicitHeight, empty.implicitHeight)
            clip: true
            boundsBehavior: Flickable.StopAtBounds

            ColumnLayout {
              id: sheet

              width: list.width
              spacing: 24

              Repeater {
                model: root.sections

                ColumnLayout {
                  id: section

                  required property var modelData

                  Layout.fillWidth: true
                  spacing: 8

                  Heading {
                    Layout.fillWidth: true
                    text: section.modelData.name
                  }

                  GridLayout {
                    id: grid

                    Layout.fillWidth: true
                    columns: sheet.width > 900 ? 2 : 1
                    columnSpacing: 32
                    rowSpacing: 0

                    Repeater {
                      model: section.modelData.binds

                      RowLayout {
                        id: row

                        required property var modelData

                        Layout.fillWidth: true
                        Layout.preferredWidth: (sheet.width - grid.columnSpacing * (grid.columns - 1)) / grid.columns
                        Layout.preferredHeight: root.rowHeight
                        spacing: 16

                        Chord {
                          Layout.preferredWidth: Theme.size(260)
                          keys: row.modelData.keys
                        }

                        Label {
                          Layout.fillWidth: true
                          text: row.modelData.text
                          elide: Text.ElideRight
                        }
                      }
                    }
                  }
                }
              }
            }

            Label {
              id: empty

              visible: root.shown.length === 0
              text: !root.loaded ? "Reading the keys…" : root.filter ? "No keys match " + root.filter : "No keys with a description"
              color: Theme.subtext
            }
          }

          Hints {
            id: hints

            Layout.fillWidth: true
            pairs: [[["Up", "Down"], "scroll"], [["Page Up", "Page Down"], "a page"], [["Home", "End"], "top and bottom"], ["", "type to filter"], ["Esc", "close"]]
          }
        }
      }
    }
  }
}
