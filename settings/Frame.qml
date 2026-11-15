import QtQuick
import QtQuick.Layouts
import qs.common

FocusScope {
  id: frame

  signal done

  property string zone: "rows"
  property int index: 0
  property bool editing: false
  property bool seeking: false
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
  readonly property var items: headers.concat(Store.headed(Store.shown.filter(row => !everyone.some(other => other.name === row.name))), heading, everyone)
  readonly property var focused: list.currentItem ? list.currentItem.item : null
  readonly property var hints: hinted()

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
    index = landing(back >= 0 ? back : index, 1);
    list.currentIndex = index;
  }

  function hinted() {
    if (chooser.visible)
      return chooser.hints;
    if (editing || list.activeFocus)
      return focused && focused.hop ? [["Tab", "next field"], ["Enter", "save"], ["Esc", "stop"]] : [["Enter", "save"], ["Esc", "cancel"]];
    const tail = [["Tab", zone === "pages" ? "rows" : "pages"], ["", "type to search"], ["Esc", Store.result === "failed" ? "dismiss" : Store.search || seeking ? "clear" : "close"]];
    if (zone === "pages")
      return [[["Up", "Down"], "pages"], ["Enter", "open"]].concat(tail);
    const item = items[index] || {};
    const own = {
      toggle: [["Space", "flip"]],
      choice: item.multiple ? [["Enter", "choose"]] : [[["Left", "Right"], "change"], ["Enter", "choose"]],
      slider: [[["Left", "Right"], "change"]],
      text: [["Enter", "edit"]],
      time: [["Enter", "edit"]],
      button: [["Enter", "open"]],
      grid: [[["Left", "Right"], "themes"], ["Enter", "apply"]],
      maker: focused && focused.open ? [[["Left", "Right"], "mode"], ["Enter", "type"]] : [["Enter", "open"]],
      pictures: [[["Left", "Right"], "pictures"], ["Enter", "show"], ["Delete", "to the Trash"]],
      about: [[["Left", "Right"], "buttons"], ["Enter", "press"]]
    }[item.kind] || [];
    return [[["Up", "Down"], "move"]].concat(own, tail);
  }

  function skipped(item) {
    return !!item && (item.kind === "everyone" || item.kind === "section");
  }

  function landing(at, step) {
    let next = Math.max(0, Math.min(items.length - 1, at));
    while (skipped(items[next]))
      next += step;
    return items[next] ? next : Math.max(0, Math.min(index, items.length - 1));
  }

  function move(step) {
    if (zone === "pages" && Store.pages.length) {
      Store.show(Store.pages[(pageIndex + step + Store.pages.length) % Store.pages.length].id);
      index = 0;
    } else if (items.length) {
      index = landing(index + step, step);
    }
  }

  function page(step) {
    if (zone === "pages")
      return turn(step);
    const from = list.itemAtIndex(index);
    let next = index;
    while (from && list.itemAtIndex(next + step) && Math.abs(list.itemAtIndex(next + step).y - from.y) <= list.height)
      next += step;
    index = landing(next === index ? index + step : next, step);
    return true;
  }

  function turn(step) {
    const starts = Store.pages.map((page, at) => at).filter(at => at === 0 || Store.pages[at - 1].group !== Store.pages[at].group);
    const here = starts.filter(at => at <= pageIndex).length - 1;
    const back = pageIndex > starts[here] ? starts[here] : starts[Math.max(0, here - 1)];
    const target = step < 0 ? back : here + 1 < starts.length ? starts[here + 1] : Store.pages.length - 1;
    Store.show(Store.pages[target].id);
    index = 0;
    return true;
  }

  function ends(step) {
    if (zone === "pages") {
      Store.show(Store.pages[step < 0 ? 0 : Store.pages.length - 1].id);
      index = 0;
    } else {
      index = landing(step < 0 ? 0 : items.length - 1, -step);
    }
    return true;
  }

  function seek() {
    Store.search = "";
    seeking = true;
    zone = "rows";
    index = 0;
    return true;
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
    if (Store.result === "failed") {
      Store.dismiss();
    } else if (Store.search || seeking) {
      Store.search = "";
      seeking = false;
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
      return fieldKey(event);
    if (event.key === Qt.Key_Slash || event.key === Qt.Key_L && event.modifiers & Qt.ControlModifier)
      return seek();
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
      [Qt.Key_Delete]: () => !!focused && !!focused.remove && focused.remove(),
      [Qt.Key_PageUp]: () => page(-1),
      [Qt.Key_PageDown]: () => page(1),
      [Qt.Key_Home]: () => ends(-1),
      [Qt.Key_End]: () => ends(1)
    };
    if (keys[event.key]) {
      seeking = seeking && event.key === Qt.Key_Escape;
      keys[event.key]();
      return true;
    }
    return typed(event);
  }

  function fieldKey(event) {
    if (event.key === Qt.Key_Escape)
      return takeKeys();
    const step = event.key === Qt.Key_Tab ? 1 : event.key === Qt.Key_Backtab ? -1 : 0;
    if (step && focused && focused.hop)
      focused.hop(step);
    return step !== 0;
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
  onIndexChanged: {
    if (skipped(items[index]))
      index = landing(index, 1);
    list.currentIndex = index;
  }
  onEditingChanged: {
    if (!editing)
      Qt.callLater(() => {
        if (!editing)
          takeKeys();
      });
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
    seeking: frame.seeking
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

    readonly property real column: Math.min(width - 64, Theme.size(880))

    anchors.left: rule.right
    anchors.right: parent.right
    anchors.top: parent.top
    anchors.bottom: status.top

    Label {
      id: title

      anchors.left: parent.left
      anchors.top: parent.top
      anchors.leftMargin: 46
      anchors.topMargin: 24
      width: content.column - 28
      text: Store.search ? "Settings matching " + Store.search : Store.pageNow.label
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
      visible: text !== ""
      text: Store.search ? "" : Store.pageNow.help || ""
      color: Theme.subtext
      wrapMode: Text.Wrap
      maximumLineCount: 2
      elide: Text.ElideRight
    }

    ListView {
      id: list

      function keepShown() {
        if (frame.skipped(frame.items[currentIndex - 1]))
          positionViewAtIndex(currentIndex - 1, ListView.Contain);
        positionViewAtIndex(currentIndex, ListView.Contain);
      }

      anchors.left: parent.left
      anchors.right: parent.right
      anchors.top: line.visible ? line.bottom : title.bottom
      anchors.bottom: keys.top
      anchors.leftMargin: 32
      anchors.rightMargin: 32
      anchors.topMargin: 12
      anchors.bottomMargin: 12
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

        width: ["grid", "maker", "pictures"].includes(entry.kind) ? ListView.view.width : content.column
        sourceComponent: ({
            about: aboutPart,
            grid: gridPart,
            maker: makerPart,
            pictures: picturesPart,
            everyone: headingPart,
            section: headingPart
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
          id: headingPart

          Item {
            readonly property bool everyone: holder.entry.kind === "everyone"

            function activate() {
              return false;
            }

            function step() {
              return false;
            }

            implicitHeight: words.implicitHeight + (everyone ? 30 : 16)

            Rectangle {
              anchors.left: parent.left
              anchors.right: parent.right
              anchors.top: parent.top
              anchors.topMargin: 10
              height: 1
              visible: parent.everyone
              color: Theme.overlay
            }

            Heading {
              id: words

              anchors.left: parent.left
              anchors.right: parent.right
              anchors.bottom: parent.bottom
              anchors.leftMargin: 14
              anchors.bottomMargin: 2
              text: parent.everyone ? "\uf023  " + "For everyone on this computer".toUpperCase() : holder.entry.label
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

    Item {
      id: keys

      anchors.left: parent.left
      anchors.bottom: parent.bottom
      anchors.leftMargin: 46
      anchors.bottomMargin: 16
      width: content.column - 28
      height: Math.min(legend.implicitHeight, legend.line * 2 + legend.spacing)
      clip: true

      Hints {
        id: legend

        readonly property real line: Math.max(0, ...Array.from(children).map(child => child.height))

        width: parent.width
        pairs: frame.hints
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
