import QtQuick

Flow {
  id: root

  property var pairs: []

  spacing: 16

  Repeater {
    model: root.pairs

    Row {
      id: pair

      required property var modelData

      spacing: 6

      Repeater {
        model: [].concat(pair.modelData[0]).filter(key => key)

        Keycap {
          required property string modelData

          anchors.verticalCenter: parent.verticalCenter
          key: modelData
        }
      }

      Label {
        anchors.verticalCenter: parent.verticalCenter
        text: pair.modelData[1]
        color: Theme.subtext
      }
    }
  }
}
