import QtQuick
import QtQuick.Layouts
import Quickshell.Io
import qs.common

Rectangle {
  id: about

  property bool current: false
  property int at: 0
  property bool credits: false
  readonly property var buttons: [
    {
      stop: "copy",
      label: "Copy the details",
      primary: true
    },
    {
      stop: "changes",
      label: "What changed",
      primary: false
    },
    {
      stop: "check",
      label: "Check the system",
      primary: false
    }
  ]
  readonly property var stops: buttons.map(button => button.stop).concat(["credits"])
  readonly property var facts: Store.about.facts || []
  readonly property string details: [[Store.about.name || "SolanumOS Chert", Store.about.version].filter(word => !!word).join(" "), "Debian " + (Store.about.debian || "13")].concat(facts.map(fact => fact.label + ": " + fact.value)).join("\n")

  function activate() {
    press(stops[at]);
    return true;
  }

  function step(direction) {
    if (at + direction < 0 || at + direction >= stops.length)
      return false;
    at += direction;
    return true;
  }

  function press(stop) {
    at = stops.indexOf(stop);
    if (stop === "copy")
      copy();
    else if (stop === "changes")
      Store.launch(["solanum", "care", "changes"]);
    else if (stop === "check")
      Store.launch(["solanum", "care", "check"]);
    else
      credits = !credits;
  }

  function copy() {
    if (Store.busy || copier.running)
      return;
    Store.said = [];
    Store.result = "";
    Store.word = "";
    copier.command = ["wl-copy", "--", details];
    copier.running = true;
  }

  implicitHeight: layout.implicitHeight + 48
  radius: Theme.radius
  color: "transparent"
  border.width: current ? 2 : 1
  border.color: current ? Theme.accent : Theme.overlay

  Process {
    id: copier

    stderr: SplitParser {
      onRead: line => Store.heard(line)
    }
    onExited: code => {
      Store.word = code === 0 ? "Copied" : "";
      Store.result = code === 0 ? "done" : "failed";
    }
  }

  ColumnLayout {
    id: layout

    anchors.left: parent.left
    anchors.right: parent.right
    anchors.top: parent.top
    anchors.margins: 24
    spacing: 24

    RowLayout {
      spacing: 24

      Item {
        Layout.preferredWidth: logo.width * logo.scale
        Layout.preferredHeight: logo.height * logo.scale

        LogoCanvas {
          id: logo

          cell: Theme.size(176) / Logo.cells
          resting: true
        }
      }

      ColumnLayout {
        Layout.fillWidth: true
        spacing: 8

        Label {
          Layout.fillWidth: true
          text: Store.about.name || "SolanumOS Chert"
          font.pixelSize: Theme.size(26)
          font.bold: true
          wrapMode: Text.Wrap
        }

        Label {
          text: Store.about.version ? "Version " + Store.about.version : ""
          color: Theme.subtext
          font.features: ({
              "tnum": 1
            })
        }

        Label {
          Layout.fillWidth: true
          text: "Built on Debian " + (Store.about.debian || "13") + ". Not part of Debian."
          wrapMode: Text.Wrap
          font.features: ({
              "tnum": 1
            })
        }

        Label {
          Layout.fillWidth: true
          text: "Hyprland on Debian, set up properly."
          color: Theme.subtext
          wrapMode: Text.Wrap
        }
      }
    }

    ColumnLayout {
      Layout.fillWidth: true
      spacing: 16

      Item {
        Layout.fillWidth: true
        implicitHeight: Theme.size(20)

        Label {
          anchors.left: parent.left
          anchors.verticalCenter: parent.verticalCenter
          text: "This computer"
          color: Theme.subtext
          font.pixelSize: Theme.size(11)
          font.bold: true
          font.capitalization: Font.AllUppercase
          font.letterSpacing: Theme.size(11) * 0.08
        }

        Label {
          anchors.right: parent.right
          anchors.verticalCenter: parent.verticalCenter
          width: Math.min(implicitWidth, parent.width * 0.6)
          visible: about.current
          text: "Left and right to move · Enter presses"
          color: Theme.subtext
          font.pixelSize: Theme.size(11)
          elide: Text.ElideLeft
          horizontalAlignment: Text.AlignRight
        }
      }

      GridLayout {
        Layout.fillWidth: true
        visible: about.facts.length > 0
        columns: 2
        columnSpacing: 32
        rowSpacing: 8

        Repeater {
          model: about.facts

          Label {
            required property var modelData
            required property int index

            Layout.row: index
            Layout.column: 0
            text: modelData.label
            color: Theme.subtext
          }
        }

        Repeater {
          model: about.facts

          Label {
            required property var modelData
            required property int index

            Layout.row: index
            Layout.column: 1
            Layout.fillWidth: true
            text: modelData.value
            elide: Text.ElideRight
            font.features: ({
                "tnum": 1
              })
          }
        }
      }
    }

    Flow {
      Layout.fillWidth: true
      spacing: 8

      Repeater {
        model: about.buttons

        Button {
          required property var modelData
          required property int index

          label: modelData.label
          primary: modelData.primary
          chosen: about.current && about.at === index
          onClicked: about.press(modelData.stop)
        }
      }
    }

    ColumnLayout {
      Layout.fillWidth: true
      spacing: 16

      Rectangle {
        id: fold

        readonly property bool chosen: about.current && about.stops[about.at] === "credits"

        Layout.fillWidth: true
        implicitHeight: foldRow.implicitHeight + 24
        radius: Theme.radius
        color: chosen ? Theme.overlay : foldPointer.containsMouse ? Theme.surface : "transparent"
        border.width: chosen ? 2 : 1
        border.color: chosen ? Theme.accent : Theme.overlay

        RowLayout {
          id: foldRow

          anchors.left: parent.left
          anchors.right: parent.right
          anchors.verticalCenter: parent.verticalCenter
          anchors.leftMargin: 14
          anchors.rightMargin: 14
          spacing: 16

          ColumnLayout {
            Layout.fillWidth: true
            spacing: 3

            Label {
              Layout.fillWidth: true
              text: "Credits and licences"
              font.bold: true
              elide: Text.ElideRight
            }

            Label {
              Layout.fillWidth: true
              text: "Who made SolanumOS, and the projects it builds on with their licences. solanum credits lists them in a terminal."
              color: Theme.subtext
              font.pixelSize: Theme.size(12)
              wrapMode: Text.Wrap
            }
          }

          Label {
            text: about.credits ? "Hide  \uf077" : "Show  \uf078"
            color: Theme.accent
          }
        }

        MouseArea {
          id: foldPointer

          anchors.fill: parent
          hoverEnabled: true
          cursorShape: Qt.PointingHandCursor
          onClicked: about.press("credits")
        }
      }

      Flow {
        Layout.fillWidth: true
        Layout.leftMargin: 14
        Layout.rightMargin: 14
        visible: about.credits
        spacing: 8

        Repeater {
          model: Store.credits

          RowLayout {
            id: credit

            required property var modelData

            width: (about.width - 84) / 2
            spacing: 8

            Label {
              Layout.maximumWidth: credit.width * 0.7
              text: credit.modelData[0]
              elide: Text.ElideRight
            }

            Label {
              Layout.fillWidth: true
              text: credit.modelData[1] || ""
              color: Theme.subtext
              elide: Text.ElideRight
            }
          }
        }
      }
    }
  }
}
