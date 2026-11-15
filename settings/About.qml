import QtQuick
import QtQuick.Layouts
import qs.common

Rectangle {
  id: about

  property bool current: false
  readonly property var facts: Store.about.facts || []

  function activate() {
    return false;
  }

  function step(direction) {
    return false;
  }

  implicitHeight: layout.implicitHeight + 48
  radius: Theme.radius
  color: "transparent"
  border.width: current ? 2 : 1
  border.color: current ? Theme.accent : Theme.overlay

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
        Layout.preferredWidth: Theme.size(176)
        Layout.preferredHeight: Theme.size(176)

        LogoCanvas {
          anchors.centerIn: parent
          cell: parent.width / Logo.cells
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

    Label {
      text: "CREDITS"
      color: Theme.subtext
      font.pixelSize: Theme.size(11)
      font.bold: true
    }

    Label {
      Layout.fillWidth: true
      text: "Who made SolanumOS, and the projects it builds on with their licences. solanum credits lists them in a terminal."
      color: Theme.subtext
      wrapMode: Text.Wrap
    }

    Flow {
      Layout.fillWidth: true
      spacing: 6

      Repeater {
        model: Store.credits

        Rectangle {
          required property var modelData

          width: (about.width - 54) / 2
          height: Theme.size(26)
          color: "transparent"

          Label {
            anchors.fill: parent
            text: "\uf004  " + parent.modelData[0] + " · " + (parent.modelData[1] || "")
            elide: Text.ElideRight
          }
        }
      }
    }
  }
}
