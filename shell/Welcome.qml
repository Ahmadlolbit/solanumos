import QtQuick
import QtQuick.Layouts
import Quickshell
import Quickshell.Hyprland
import Quickshell.Io
import Quickshell.Wayland
import qs.common

Scope {
  id: root

  readonly property string stateDir: Quickshell.env("HOME") + "/.local/state/solanum"
  readonly property var keys: [["Super + Space", "Apps"], ["Super + Alt + Space", "Menu"], ["Super + K", "All the keys"], ["Super + Return", "Terminal"], ["Super + Ctrl + T", "Themes"]]
  property bool open: false
  property string tip: ""
  readonly property var screen: Quickshell.screens.find(item => Hyprland.focusedMonitor && item.name === Hyprland.focusedMonitor.name) || Quickshell.screens[0]

  function show() {
    tip = "";
    tipper.running = true;
    open = true;
    settings.forceActiveFocus();
  }

  function close() {
    open = false;
    Quickshell.execDetached(["sh", "-c", "mkdir -p \"$1\" && touch \"$1/welcome-seen\"", "sh", stateDir]);
  }

  function run(command) {
    close();
    Quickshell.execDetached(["sh", "-c", command]);
  }

  function step(forward) {
    const from = content.Window.activeFocusItem;
    const next = from ? from.nextItemInFocusChain(forward) : null;
    if (next)
      next.forceActiveFocus(forward ? Qt.TabFocusReason : Qt.BacktabFocusReason);
  }

  onOpenChanged: {
    if (open)
      Keyboard.used = false;
  }

  FileView {
    id: seen

    path: root.stateDir + "/welcome-seen"
    printErrors: false
    onLoadFailed: later.start()
  }

  Process {
    id: tipper

    command: ["solanum-tip", "welcome"]
    stdout: StdioCollector {
      onStreamFinished: root.tip = text.trim()
    }
  }

  Timer {
    id: later

    interval: 2500
    onTriggered: root.show()
  }

  IpcHandler {
    target: "welcome"

    function open(): void {
      root.show();
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
    WlrLayershell.namespace: "solanum-welcome"
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
        else if (event.key === Qt.Key_Left || event.key === Qt.Key_Right)
          root.step(event.key === Qt.Key_Right);
        else
          return;
        event.accepted = true;
      }

      Fade on opacity {}

      Rectangle {
        anchors.centerIn: parent
        width: column.implicitWidth + 64
        height: column.implicitHeight + 56
        color: Theme.panel
        border.color: Theme.accent
        border.width: 2
        radius: Theme.radius

        ColumnLayout {
          id: column

          anchors.centerIn: parent
          spacing: 18

          Item {
            Layout.alignment: Qt.AlignHCenter
            Layout.preferredWidth: Theme.size(176)
            Layout.preferredHeight: Theme.size(176)

            Loader {
              active: root.open || content.opacity > 0

              sourceComponent: LogoCanvas {
                cell: Theme.size(176) / Logo.cells
                resting: true
              }
            }
          }

          Label {
            Layout.alignment: Qt.AlignHCenter
            text: "Welcome. These keys get you everywhere:"
            color: Theme.subtext
          }

          ColumnLayout {
            Layout.alignment: Qt.AlignHCenter
            spacing: 8

            Repeater {
              model: root.keys

              RowLayout {
                id: shortcut

                required property var modelData

                spacing: 24

                Item {
                  Layout.preferredWidth: Theme.size(200)
                  Layout.preferredHeight: chord.implicitHeight

                  Chord {
                    id: chord

                    anchors.right: parent.right
                    keys: shortcut.modelData[0]
                  }
                }

                Label {
                  text: shortcut.modelData[1]
                }
              }
            }
          }

          RowLayout {
            Layout.alignment: Qt.AlignHCenter
            Layout.topMargin: 6
            spacing: 8

            Button {
              id: settings

              primary: true
              label: "Open Settings"
              onClicked: root.run("solanum settings")
            }

            Button {
              label: "Pick a theme"
              onClicked: root.run("qs -p /usr/share/solanum/shell ipc call picker themes")
            }

            Button {
              label: "Pick a background"
              onClicked: root.run("qs -p /usr/share/solanum/shell ipc call picker backgrounds")
            }

            Button {
              label: "Project page"
              onClicked: root.run("xdg-open https://github.com/Ahmadlolbit/solanumos")
            }

            Button {
              label: "Close"
              onClicked: root.close()
            }
          }

          Label {
            Layout.alignment: Qt.AlignHCenter
            Layout.maximumWidth: Theme.size(640)
            visible: root.tip !== ""
            text: "Tip: " + root.tip
            wrapMode: Text.WordWrap
            horizontalAlignment: Text.AlignHCenter
          }

          Label {
            Layout.alignment: Qt.AlignHCenter
            text: "Menu › Learn › Welcome brings this back."
            color: Theme.subtext
          }
        }
      }
    }
  }
}
