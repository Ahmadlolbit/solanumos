import QtQuick
import QtQuick.Layouts
import qs.common

Rectangle {
  id: about

  property bool current: false

  function activate() {
    return false;
  }

  function step(direction) {
    return false;
  }

  implicitHeight: layout.implicitHeight + 40
  radius: Theme.radius
  color: "transparent"
  border.width: current ? 2 : 1
  border.color: current ? Theme.accent : Theme.overlay

  ColumnLayout {
    id: layout

    anchors.left: parent.left
    anchors.right: parent.right
    anchors.top: parent.top
    anchors.margins: 20
    spacing: 18

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
        spacing: 6

        Label {
          text: Store.about.name || "SolanumOS Chert"
          font.pixelSize: Theme.size(26)
          font.bold: true
        }

        Label {
          text: Store.about.version ? "Version " + Store.about.version : ""
          color: Theme.subtext
        }

        Label {
          Layout.fillWidth: true
          text: "Built on Debian " + (Store.about.debian || "13") + ". Not part of Debian."
          wrapMode: Text.Wrap
        }

        Label {
          Layout.fillWidth: true
          text: "Hyprland on Debian, set up properly."
          color: Theme.subtext
          wrapMode: Text.Wrap
        }
      }
    }

    Label {
      text: "THIS COMPUTER"
      color: Theme.subtext
      font.pixelSize: Theme.size(11)
      font.bold: true
    }

    GridLayout {
      Layout.fillWidth: true
      columns: 2
      columnSpacing: 24
      rowSpacing: 6

      Repeater {
        model: Store.about.facts || []

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
        model: Store.about.facts || []

        Label {
          required property var modelData
          required property int index

          Layout.row: index
          Layout.column: 1
          Layout.fillWidth: true
          text: modelData.value
          elide: Text.ElideRight
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

          width: (about.width - 52) / 2
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
