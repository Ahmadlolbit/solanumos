import QtQuick
import qs.common

Tile {
  id: grid

  property int index: Math.max(0, Store.themes.findIndex(item => item.name === Store.current))
  readonly property int across: Math.max(1, Math.floor((width - 32 + 10) / (Theme.size(176) + 10)))

  signal pointed

  function activate() {
    const theme = Store.themes[index];
    if (theme)
      Store.act("theme", ["solanum-theme-set", theme.name], "themes");
    return !!theme;
  }

  function step(direction) {
    if (!Store.themes.length)
      return false;
    index = (index + direction + Store.themes.length) % Store.themes.length;
    return true;
  }

  title: "Themes"
  hint: current ? "Left and right to move · Enter applies" : Store.themes.length + " themes, yours among them"

  Flow {
    width: parent.width
    spacing: 10

    Repeater {
      model: Store.themes

      Rectangle {
        id: theme

        required property var modelData
        required property int index
        readonly property bool picked: index === grid.index && grid.current

        width: Theme.size(176)
        height: preview.height + name.height + 14
        radius: Theme.radius
        color: modelData.name === Store.current ? Theme.overlay : Theme.surface
        border.width: picked ? 2 : modelData.name === Store.current ? 1 : 0
        border.color: picked ? Theme.accent : Theme.subtext

        Image {
          id: preview

          anchors.left: parent.left
          anchors.right: parent.right
          anchors.top: parent.top
          anchors.margins: 4
          height: width * 10 / 16
          source: theme.modelData.image ? "file://" + theme.modelData.image.split("/").map(encodeURIComponent).join("/") : ""
          fillMode: Image.PreserveAspectCrop
          asynchronous: true
          sourceSize.width: 320
          sourceSize.height: 200
        }

        Label {
          id: name

          anchors.left: parent.left
          anchors.right: parent.right
          anchors.top: preview.bottom
          anchors.margins: 6
          text: (theme.modelData.name === Store.current ? "\uf00c " : "") + theme.modelData.label
          elide: Text.ElideRight
          font.pixelSize: Theme.size(12)
        }

        MouseArea {
          anchors.fill: parent
          cursorShape: Qt.PointingHandCursor
          onClicked: {
            grid.pointed();
            grid.index = theme.index;
            grid.activate();
          }
        }
      }
    }
  }
}
