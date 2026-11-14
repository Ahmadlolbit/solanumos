import QtQuick
import qs.common

Item {
  id: sidebar

  property bool active: false

  signal picked(string id)

  Rectangle {
    anchors.fill: parent
    color: Theme.surface
  }

  Rectangle {
    id: search

    anchors.left: parent.left
    anchors.right: parent.right
    anchors.top: parent.top
    anchors.margins: 16
    height: Theme.size(32)
    radius: Theme.radius
    color: Theme.base
    border.width: Store.search ? 2 : 1
    border.color: Store.search ? Theme.accent : Theme.overlay

    Label {
      anchors.fill: parent
      anchors.leftMargin: 10
      anchors.rightMargin: 10
      text: Store.search ? "\uf002  " + Store.search : "\uf002  Type to search"
      color: Store.search ? Theme.text : Theme.subtext
      elide: Text.ElideLeft
    }
  }

  ListView {
    id: pages

    anchors.left: parent.left
    anchors.right: parent.right
    anchors.top: search.bottom
    anchors.bottom: parent.bottom
    anchors.margins: 10
    anchors.topMargin: 12
    clip: true
    spacing: 2
    model: Store.pages
    currentIndex: Math.max(0, Store.pages.findIndex(item => item.id === Store.page))
    highlightMoveDuration: 0
    boundsBehavior: Flickable.StopAtBounds
    onCurrentIndexChanged: positionViewAtIndex(currentIndex, ListView.Contain)

    delegate: Choice {
      required property var modelData

      width: ListView.view.width
      icon: modelData.glyph
      label: modelData.label
      selected: modelData.id === Store.page && !Store.search
      border.width: selected && sidebar.active ? 2 : 0
      border.color: Theme.accent
      onClicked: sidebar.picked(modelData.id)
    }
  }
}
