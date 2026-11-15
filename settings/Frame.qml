import QtQuick
import QtQuick.Layouts
import qs.common

FocusScope {
  id: frame

  signal done

  property string zone: "rows"
  property int index: 0
  property bool editing: false
  readonly property int pageIndex: Math.max(0, Store.pages.findIndex(item => item.id === Store.page))
  readonly property var headers: Store.search ? [] : ({
      about: [
        {
          kind: "about",
          name: "header.about"
        }
      ],
      themes: [
        {
          kind: "grid",
          name: "header.grid"
        },
        {
          kind: "maker",
          name: "header.maker"
        }
      ],
      wallpaper: [
        {
          kind: "pictures",
          name: "header.theme",
          every: false
        },
        {
          kind: "pictures",
          name: "header.every",
          every: true
        }
      ]
    })[Store.page] || []
  readonly property var everyone: Store.search ? [] : Store.shown.filter(row => row.whose === "computer" && !!row.twin).sort((one, other) => Store.rows.findIndex(row => row.name === one.twin) - Store.rows.findIndex(row => row.name === other.twin))
  readonly property var heading: everyone.length ? [
      {
        kind: "everyone",
        name: "header.everyone"
      }
    ] : []
  readonly property var items: headers.concat(Store.shown.filter(row => !everyone.some(other => other.name === row.name)), heading, everyone)
  readonly property var focused: list.currentItem ? list.currentItem.item : null

  function sync() {
    const kept = index < listed.count ? listed.get(index).name : "";
    const unique = items.filter((item, at) => items.findIndex(other => other.name === item.name) === at);
    const names = unique.map(item => item.name);
    for (let at = listed.count - 1; at >= 0; at--) {
      if (!names.includes(listed.get(at).name))
        listed.remove(at);
    }
    unique.forEach((item, at) => {
      const json = JSON.stringify(item);
      let from = at;
      while (from < listed.count && listed.get(from).name !== item.name)
        from++;
      if (from === listed.count)
        listed.insert(at, {
          name: item.name,
          json: json
        });
      else if (from !== at)
        listed.move(from, at, 1);
      if (listed.get(at).json !== json)
        listed.setProperty(at, "json", json);
    });
    if (listed.count > names.length)
      listed.remove(names.length, listed.count - names.length);
    const back = names.indexOf(kept);
    index = back >= 0 ? back : Math.max(0, Math.min(index, names.length - 1));
    list.currentIndex = index;
  }

  function move(step) {
    if (zone === "pages" && Store.pages.length) {
      Store.show(Store.pages[(pageIndex + step + Store.pages.length) % Store.pages.length].id);
      index = 0;
    } else if (items.length) {
      let next = Math.max(0, Math.min(items.length - 1, index + step));
      if (items[next].kind === "everyone")
        next = items[next + step] ? next + step : index;
      index = next;
    }
  }

  function typed(event) {
    if (event.key === Qt.Key_Backspace && Store.search) {
      Store.search = Store.search.slice(0, -1);
    } else if (/^[\w .,+-]$/.test(event.text)) {
      Store.search += event.text;
      zone = "rows";
    } else {
      return false;
    }
    index = 0;
    return true;
  }

  function leave() {
    if (Store.search) {
      Store.search = "";
      index = 0;
    } else {
      done();
    }
    return true;
  }

  function confirm() {
    if (zone === "pages") {
      zone = "rows";
      index = 0;
      return true;
    }
    if (Store.search && focused && focused.row && focused.row.page !== Store.page && !focused.row.command) {
      const name = focused.row.name;
      Store.show(focused.row.page);
      index = Math.max(0, items.findIndex(item => item.name === name));
      return true;
    }
    return !!focused && focused.activate();
  }

  function sideways(step) {
    if (zone === "rows" && focused && focused.step(step))
      return true;
    zone = step < 0 ? "pages" : "rows";
    return true;
  }

  function key(event) {
    if (chooser.visible)
      return chooser.key(event);
    if (editing || list.activeFocus)
      return event.key === Qt.Key_Escape ? takeKeys() : event.key === Qt.Key_Tab || event.key === Qt.Key_Backtab;
    const keys = {
      [Qt.Key_Escape]: () => leave(),
      [Qt.Key_Tab]: () => flip(),
      [Qt.Key_Backtab]: () => flip(),
      [Qt.Key_Up]: () => move(-1),
      [Qt.Key_Down]: () => move(1),
      [Qt.Key_Left]: () => sideways(-1),
      [Qt.Key_Right]: () => sideways(1),
      [Qt.Key_Return]: () => confirm(),
      [Qt.Key_Enter]: () => confirm(),
      [Qt.Key_Space]: () => Store.search ? typed(event) : zone === "rows" && !!focused && focused.activate(),
      [Qt.Key_Delete]: () => !!focused && !!focused.remove && focused.remove()
    };
    if (keys[event.key]) {
      keys[event.key]();
      return true;
    }
    return typed(event);
  }

  function flip() {
    zone = zone === "pages" ? "rows" : "pages";
    return true;
  }

  function takeKeys() {
    if (focused && "editing" in focused)
      focused.editing = false;
    editing = false;
    list.focus = false;
    frame.forceActiveFocus();
    return true;
  }

  function point(at) {
    if (at !== index && editing)
      takeKeys();
    index = at;
  }

  function choose(row) {
    chooser.open(row);
  }

  focus: true
  Component.onCompleted: sync()
  onItemsChanged: sync()
  onIndexChanged: list.currentIndex = index
  onEditingChanged: {
    if (!editing)
      Qt.callLater(takeKeys);
  }
  Keys.onPressed: event => {
    if (frame.key(event))
      event.accepted = true;
  }

  Connections {
    target: Store

    function onPageChanged() {
      if (frame.zone !== "pages")
        frame.index = 0;
    }
  }

  ListModel {
    id: listed
  }

  Rectangle {
    anchors.fill: parent
    color: Theme.base
  }

  Sidebar {
    id: sidebar

    anchors.left: parent.left
    anchors.top: parent.top
    anchors.bottom: parent.bottom
    width: Theme.size(280)
    active: frame.zone === "pages"
    onPicked: id => {
      Store.show(id);
      frame.zone = "rows";
      frame.index = 0;
      frame.takeKeys();
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

    anchors.left: rule.right
    anchors.right: parent.right
    anchors.top: parent.top
    anchors.bottom: status.top

    Label {
      id: title

      anchors.left: parent.left
      anchors.right: parent.right
      anchors.top: parent.top
      anchors.margins: 24
      text: Store.search ? "Settings matching " + Store.search : Store.pageNow.label
      font.pixelSize: Theme.size(22)
      font.bold: true
      elide: Text.ElideRight
    }

    ListView {
      id: list

      function keepShown() {
        const above = frame.items[currentIndex - 1];
        if (above && above.kind === "everyone")
          positionViewAtIndex(currentIndex - 1, ListView.Contain);
        positionViewAtIndex(currentIndex, ListView.Contain);
      }

      anchors.left: parent.left
      anchors.right: parent.right
      anchors.top: title.bottom
      anchors.bottom: parent.bottom
      anchors.margins: 16
      anchors.topMargin: 12
      clip: true
      spacing: 8
      cacheBuffer: 100000
      model: listed
      highlightMoveDuration: 0
      boundsBehavior: Flickable.StopAtBounds
      onCurrentIndexChanged: keepShown()
      onContentHeightChanged: keepShown()

      delegate: Loader {
        id: holder

        required property string name
        required property string json
        required property int index
        readonly property var entry: JSON.parse(json)

        width: ListView.view.width
        sourceComponent: ({
            about: aboutPart,
            grid: gridPart,
            maker: makerPart,
            pictures: picturesPart,
            everyone: everyonePart
          })[entry.kind] || rowPart

        Component {
          id: rowPart

          RowItem {
            row: holder.entry
            current: holder.index === frame.index && frame.zone === "rows"
            searching: !!Store.search
            grouped: frame.everyone.some(row => row.name === holder.name)
            onChosen: frame.choose(row)
            onEditingChanged: frame.editing = editing
            onPointed: frame.point(holder.index)
          }
        }

        Component {
          id: everyonePart

          Item {
            function activate() {
              return false;
            }

            function step() {
              return false;
            }

            implicitHeight: everyoneTitle.implicitHeight + 30

            Rectangle {
              anchors.left: parent.left
              anchors.right: parent.right
              anchors.top: parent.top
              anchors.topMargin: 10
              height: 1
              color: Theme.overlay
            }

            Label {
              id: everyoneTitle

              anchors.left: parent.left
              anchors.bottom: parent.bottom
              anchors.leftMargin: 14
              anchors.bottomMargin: 2
              text: "\uf023  " + "For everyone on this computer".toUpperCase()
              color: Theme.subtext
              font.pixelSize: Theme.size(11)
              font.bold: true
            }
          }
        }

        Component {
          id: aboutPart

          About {
            current: holder.index === frame.index && frame.zone === "rows"
          }
        }

        Component {
          id: gridPart

          ThemeGrid {
            current: holder.index === frame.index && frame.zone === "rows"
            onPointed: frame.point(holder.index)
          }
        }

        Component {
          id: makerPart

          Maker {
            current: holder.index === frame.index && frame.zone === "rows"
            onEditingChanged: frame.editing = editing
            onPointed: frame.point(holder.index)
          }
        }

        Component {
          id: picturesPart

          Pictures {
            every: holder.entry.every
            current: holder.index === frame.index && frame.zone === "rows"
            onPointed: frame.point(holder.index)
          }
        }
      }
    }

    Label {
      anchors.centerIn: list
      visible: frame.items.length === 0
      text: Store.search ? "Nothing matches " + Store.search : Store.rows.length ? "Nothing here yet" : "Loading"
      color: Theme.subtext
    }

    Chooser {
      id: chooser

      anchors.fill: parent
      onClosed: frame.takeKeys()
    }
  }

  Status {
    id: status

    anchors.left: rule.right
    anchors.right: parent.right
    anchors.bottom: parent.bottom
  }
}
