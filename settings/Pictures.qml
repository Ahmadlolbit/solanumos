import QtQuick
import qs.common

Tile {
  id: box

  property bool every: false
  property int index: 0
  readonly property string folder: every ? Store.yours : Store.yours + Store.current + "/"
  readonly property var mine: Store.pictures.filter(path => path.startsWith(folder) && !path.slice(folder.length).includes("/"))
  readonly property var tiles: mine.concat(["add"])

  signal pointed

  function activate() {
    const picked = tiles[Math.min(index, tiles.length - 1)];
    if (picked === "add")
      Store.act("add", every ? ["solanum-background", "add", "--every"] : ["solanum-background", "add"], "pictures");
    else
      Store.act("show", ["solanum-background", picked], "pictures");
    return true;
  }

  function step(direction) {
    index = (index + direction + tiles.length) % tiles.length;
    return true;
  }

  function remove() {
    const picked = tiles[Math.min(index, tiles.length - 1)];
    if (picked === "add")
      return false;
    Store.act("remove", ["solanum-background", "remove", picked], "pictures");
    return true;
  }

  title: every ? "Yours for every theme" : "Yours for " + (Store.current ? Store.currentLabel : "this theme")
  hint: current ? "Enter shows it or adds pictures · Delete moves one to the Trash" : every ? "~/.config/solanum/backgrounds" : "~/.config/solanum/backgrounds/" + Store.current

  Flow {
    width: parent.width
    spacing: 10

    Repeater {
      model: box.tiles

      Rectangle {
        id: picture

        required property string modelData
        required property int index
        readonly property bool picked: index === box.index && box.current

        width: Theme.size(150)
        height: width * 10 / 16
        radius: Theme.radius
        color: Theme.surface
        border.width: picked ? 2 : 0
        border.color: Theme.accent
        clip: true

        Image {
          anchors.fill: parent
          anchors.margins: picture.picked ? 2 : 0
          visible: picture.modelData !== "add"
          source: picture.modelData !== "add" ? "file://" + picture.modelData.split("/").map(encodeURIComponent).join("/") : ""
          fillMode: Image.PreserveAspectCrop
          asynchronous: true
          sourceSize.width: 300
          sourceSize.height: 188
        }

        Label {
          anchors.centerIn: parent
          visible: picture.modelData === "add"
          text: "\uf067  Add pictures"
          color: Theme.accent
        }

        Rectangle {
          anchors.right: parent.right
          anchors.top: parent.top
          anchors.margins: 6
          width: Theme.size(24)
          height: width
          radius: Theme.radius
          visible: picture.modelData !== "add"
          color: Qt.alpha(Theme.base, 0.85)

          Label {
            anchors.centerIn: parent
            text: "\uf1f8"
            color: trash.containsMouse ? Theme.critical : Theme.text
          }

          MouseArea {
            id: trash

            anchors.fill: parent
            hoverEnabled: true
            cursorShape: Qt.PointingHandCursor
            onClicked: {
              box.pointed();
              box.index = picture.index;
              box.remove();
            }
          }
        }

        MouseArea {
          anchors.fill: parent
          z: -1
          cursorShape: Qt.PointingHandCursor
          onClicked: {
            box.pointed();
            box.index = picture.index;
            box.activate();
          }
        }
      }
    }
  }
}
