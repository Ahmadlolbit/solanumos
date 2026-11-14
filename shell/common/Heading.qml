import QtQuick

Item {
  id: root

  property string text: ""
  property string note: ""

  implicitWidth: title.implicitWidth + (side.visible ? side.implicitWidth + 16 : 0)
  implicitHeight: Theme.size(20)

  Label {
    id: title

    anchors.left: parent.left
    anchors.verticalCenter: parent.verticalCenter
    width: Math.min(implicitWidth, root.width - (side.visible ? side.width + 16 : 0))
    text: root.text
    color: Theme.subtext
    elide: Text.ElideRight
    font.family: Theme.display
    font.pixelSize: Theme.size(11)
    font.bold: Theme.displayBold
    font.capitalization: Font.AllUppercase
    font.letterSpacing: Theme.size(11) * 0.08
  }

  Label {
    id: side

    anchors.right: parent.right
    anchors.verticalCenter: parent.verticalCenter
    width: Math.min(implicitWidth, root.width * 0.7)
    visible: root.note !== ""
    text: root.note
    color: Theme.subtext
    elide: Text.ElideLeft
    horizontalAlignment: Text.AlignRight
    font.pixelSize: Theme.size(11)
  }
}
