import QtQuick
import QtQuick.Layouts
import qs.common
import "." as Care

FocusScope {
  id: frame

  property string zone: "page"
  property bool asking: false
  readonly property Item shown: pages.item
  readonly property bool following: zone === "page" && !!shown && !!shown.working
  readonly property var hints: asking ? [[["Left", "Right"], "pick"], ["Enter", "answer"]] : zone === "pages" ? [[["Up", "Down"], "page"], ["Enter", "open"], ["Tab", "page"], ["Esc", "close"]] : (shown && shown.hints ? shown.hints : []).concat([[following ? "Shift+Tab" : "Tab", "pages"], ["Esc", "close"]])

  signal done

  function takeKeys() {
    if (asking)
      keep.forceActiveFocus();
    else if (zone === "page" && shown)
      shown.enter();
    else
      sidebarKeys();
    return true;
  }

  function sidebarKeys() {
    pages.focus = false;
    if (shown)
      shown.focus = false;
    frame.forceActiveFocus();
  }

  function ask() {
    asking = true;
    Store.holding = true;
    takeKeys();
  }

  function answered(stop) {
    asking = false;
    if (stop)
      Store.task.stop();
    Store.holding = false;
    return takeKeys();
  }

  function leave() {
    if (asking)
      return answered(false);
    if (Store.running)
      ask();
    else
      done();
    return true;
  }

  function turn(step) {
    const at = Store.pages.findIndex(item => item.id === Store.page);
    Store.show(Store.pages[(at + step + Store.pages.length) % Store.pages.length].id);
    return true;
  }

  function sideways(key) {
    if (key !== Qt.Key_Left && key !== Qt.Key_Right)
      return false;
    (keep.activeFocus ? stop : keep).forceActiveFocus();
    return true;
  }

  function key(event) {
    if (event.key === Qt.Key_Escape)
      return leave();
    if (asking)
      return sideways(event.key);
    if (event.key === Qt.Key_Backtab || event.key === Qt.Key_Tab && !following) {
      zone = zone === "pages" ? "page" : "pages";
      return takeKeys();
    }
    if (zone !== "pages")
      return false;
    if (event.key === Qt.Key_Up || event.key === Qt.Key_Down)
      return turn(event.key === Qt.Key_Up ? -1 : 1);
    if (event.key === Qt.Key_Return || event.key === Qt.Key_Enter) {
      zone = "page";
      return takeKeys();
    }
    return false;
  }

  Keys.onPressed: event => event.accepted = frame.key(event)

  Connections {
    target: Store

    function onPageChanged() {
      Qt.callLater(frame.takeKeys);
    }

    function onRunningChanged() {
      if (!Store.running && frame.asking)
        frame.answered(false);
    }
  }

  Rectangle {
    anchors.fill: parent
    color: Theme.base
  }

  Rectangle {
    id: sidebar

    anchors.left: parent.left
    anchors.top: parent.top
    anchors.bottom: parent.bottom
    width: Theme.size(240)
    color: Theme.surface

    Label {
      id: name

      anchors.left: parent.left
      anchors.top: parent.top
      anchors.margins: 16
      anchors.leftMargin: 18
      text: "Care"
      font.family: Theme.display
      font.bold: Theme.displayBold
      font.pixelSize: Theme.size(16)
    }

    Column {
      anchors.left: parent.left
      anchors.right: parent.right
      anchors.top: name.bottom
      anchors.margins: 10
      anchors.topMargin: 16
      spacing: 2

      Repeater {
        model: Store.pages

        Choice {
          required property var modelData

          width: parent.width
          icon: modelData.glyph
          label: modelData.label
          selected: modelData.id === Store.page
          activeFocusOnTab: false
          border.width: selected && frame.zone === "pages" && !frame.asking ? Theme.focusBorder : 0
          border.color: Theme.accent
          onClicked: {
            Store.show(modelData.id);
            frame.zone = "page";
            frame.takeKeys();
          }
        }
      }
    }
  }

  Rectangle {
    id: rule

    anchors.left: sidebar.right
    anchors.top: parent.top
    anchors.bottom: parent.bottom
    width: 1
    color: Theme.overlay
  }

  Item {
    id: content

    readonly property real column: Math.min(width - 64, Theme.size(880))

    anchors.left: rule.right
    anchors.right: parent.right
    anchors.top: parent.top
    anchors.bottom: parent.bottom

    Label {
      id: title

      x: Math.round((content.width - content.column) / 2)
      anchors.top: parent.top
      anchors.topMargin: 24
      width: content.column
      text: Store.pageNow.label
      font.pixelSize: Theme.size(22)
      font.bold: true
      elide: Text.ElideRight
    }

    Label {
      id: line

      anchors.left: title.left
      anchors.right: title.right
      anchors.top: title.bottom
      anchors.topMargin: 4
      text: Store.pageNow.help
      color: Theme.subtext
      elide: Text.ElideRight
    }

    Loader {
      id: pages

      anchors.left: title.left
      anchors.right: title.right
      anchors.top: line.bottom
      anchors.bottom: question.visible ? question.top : keys.top
      anchors.topMargin: 24
      anchors.bottomMargin: 16
      sourceComponent: ({
          updates: updatesPart,
          check: checkPart,
          changes: changesPart,
          restore: restorePart,
          history: historyPart,
          space: spacePart,
          undo: undoPart,
          report: reportPart
        })[Store.pageNow.id]
      onLoaded: Qt.callLater(frame.takeKeys)

      Component {
        id: updatesPart

        Care.Updates {}
      }

      Component {
        id: checkPart

        Care.Check {}
      }

      Component {
        id: changesPart

        Care.Changes {}
      }

      Component {
        id: restorePart

        Care.Restore {}
      }

      Component {
        id: historyPart

        Care.History {}
      }

      Component {
        id: spacePart

        Care.Space {}
      }

      Component {
        id: undoPart

        Care.Undo {}
      }

      Component {
        id: reportPart

        Care.Report {}
      }
    }

    Rectangle {
      id: question

      anchors.left: title.left
      anchors.right: title.right
      anchors.bottom: keys.top
      anchors.bottomMargin: 16
      visible: frame.asking
      height: asked.implicitHeight + 32
      color: Theme.surface
      radius: Theme.radius
      border.width: 1
      border.color: Theme.overlay

      RowLayout {
        id: asked

        anchors.fill: parent
        anchors.margins: 16
        spacing: 16

        Label {
          Layout.fillWidth: true
          text: Store.stopWords + " If you keep it running, this window stays open on it."
          wrapMode: Text.Wrap
        }

        Button {
          id: keep

          label: "Keep it running"
          primary: true
          onClicked: frame.answered(false)
        }

        Button {
          id: stop

          label: "Stop"
          onClicked: frame.answered(true)
        }
      }
    }

    Hints {
      id: keys

      anchors.left: title.left
      anchors.right: title.right
      anchors.bottom: parent.bottom
      anchors.bottomMargin: 16
      pairs: frame.hints
    }
  }
}
