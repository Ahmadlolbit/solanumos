import QtQuick
import qs.common

FocusScope {
  id: frame

  property string zone: "rows"
  property int index: 0
  property var asking: null
  property bool adding: false
  readonly property bool searching: Store.search !== ""
  readonly property bool watching: Store.page === "running" && !searching
  readonly property var items: listed()
  readonly property var focused: items[index] || null
  readonly property var hints: hinted()

  signal done

  function section(label) {
    return {
      kind: "section",
      label: label
    };
  }

  function note(label) {
    return {
      kind: "note",
      label: label
    };
  }

  function appItem(app) {
    return {
      kind: "app",
      id: app.id,
      name: app.name,
      detail: app.from,
      installed: app.installed
    };
  }

  function packageItem(found) {
    return {
      kind: "pkg",
      id: found.name,
      name: found.name,
      detail: found.description,
      installed: found.installed
    };
  }

  function appItems() {
    if (!Store.appsRead)
      return [note("Reading the apps")];
    const found = [];
    Store.apps.forEach((app, at) => {
      if (at === 0 || Store.apps[at - 1].group !== app.group)
        found.push(section(app.group));
      found.push(appItem(app));
    });
    return found;
  }

  function packageItems(text) {
    if (!Store.packagesRead)
      return [note("Reading every Debian package")];
    const found = Store.packageMatches(text).map(packageItem);
    if (found.length > 0)
      return found;
    if (!text && !Store.installedOnly)
      return [note("Type to search " + Store.packages.length.toLocaleString(Qt.locale("en_GB"), "f", 0) + " packages")];
    return [note(text ? "No package matches " + text : "No package is installed")];
  }

  function webItems() {
    const add = {
      kind: "add",
      label: "Add a web app",
      detail: "A name, its address, and an icon if you have one"
    };
    return [add].concat(Store.webapps.map(web => ({
          kind: "web",
          id: web.name,
          name: web.name,
          detail: web.address
        })));
  }

  function searchItems() {
    const apps = Store.appMatches(Store.search).map(appItem);
    return [section("Apps")].concat(apps.length > 0 ? apps : [note("No app matches " + Store.search)], [section("Debian packages")], packageItems(Store.search));
  }

  function listed() {
    if (searching)
      return searchItems();
    if (Store.page === "apps")
      return appItems();
    if (Store.page === "packages")
      return [
        {
          kind: "filter",
          label: "Only installed packages",
          detail: Store.installedOnly ? "On: Enter shows every package" : "Off: Enter shows only the installed ones",
          installed: Store.installedOnly
        }
      ].concat(packageItems(""));
    return Store.page === "webapps" ? webItems() : [];
  }

  function hinted() {
    if (asking)
      return [["Enter", "yes"], ["Esc", "no"]];
    if (adding)
      return [["Tab", "next field"], ["Enter", "add"], ["Esc", "cancel"]];
    if (zone === "pages")
      return [[["Up", "Down"], "page"], [["Enter", "Tab"], "open"], ["Esc", "close"]];
    if (watching)
      return [["Esc", "pages"]];
    const marking = focused && ["app", "pkg"].includes(focused.kind);
    return [[["Up", "Down"], "move"], ["Enter", Store.marked > 0 ? "apply" : marking ? (focused.installed ? "remove" : "install") : "choose"]].concat(marking && !searching ? [["Space", "mark"]] : [], [["Type", "search"], ["Esc", searching ? "clear" : "close"]]);
  }

  function selectable(item) {
    return !!item && !["section", "note"].includes(item.kind);
  }

  function move(step) {
    let at = index + step;
    while (at >= 0 && at < items.length && !selectable(items[at]))
      at += step;
    if (at >= 0 && at < items.length)
      index = at;
    return true;
  }

  function first() {
    index = 0;
    if (!selectable(items[0]))
      move(1);
  }

  function seen() {
    if (!Store.wanted || !Store.appsRead)
      return;
    const at = items.findIndex(item => item.kind === "app" && item.id === Store.wanted);
    Store.wanted = "";
    if (at >= 0)
      index = at;
  }

  function turn(step) {
    const at = Store.pages.findIndex(page => page.id === Store.page);
    Store.show(Store.pages[Math.max(0, Math.min(Store.pages.length - 1, at + step))].id);
    return true;
  }

  function single(item) {
    const kind = item.kind === "app" ? "app-" : "pkg-";
    return [
      {
        job: kind + (item.installed ? "remove" : "install"),
        args: [item.id],
        label: (item.installed ? "Remove " : "Install ") + item.name
      }
    ];
  }

  function ask(text, yes) {
    asking = {
      text: text,
      yes: yes
    };
    return true;
  }

  function askChanges(list) {
    const text = list.length === 1 ? list[0].label + "?" : "Apply " + list.length + " changes: " + list.map(change => change.label.charAt(0).toLowerCase() + change.label.slice(1)).join("; ") + "?";
    return ask(text, () => Store.apply(list));
  }

  function activate() {
    const item = focused;
    if (!selectable(item))
      return true;
    if (item.kind === "filter")
      Store.installedOnly = !Store.installedOnly;
    else if (item.kind === "add")
      adding = true;
    else if (item.kind === "web")
      ask("Remove the web app " + item.name + "?", () => Store.removeWebApp(item.name));
    else
      askChanges(Store.marked > 0 ? Store.changes() : single(item));
    return true;
  }

  function markFocused() {
    const item = focused;
    if (item && ["app", "pkg"].includes(item.kind))
      Store.mark(item.kind, item.id, item.name, item.installed);
    else if (item && item.kind === "filter")
      Store.installedOnly = !Store.installedOnly;
    return true;
  }

  function answer(yes) {
    const then = asking ? asking.yes : null;
    asking = null;
    if (yes && then)
      then();
    return true;
  }

  function askStop() {
    asking = null;
    adding = false;
    Store.search = "";
    Store.show("running");
    zone = "rows";
    view.askStop();
  }

  function leave() {
    if (asking)
      return answer(false);
    if (adding) {
      adding = false;
      return true;
    }
    if (searching) {
      Store.search = "";
      first();
    } else if (zone === "rows" && watching) {
      zone = "pages";
    } else if (Store.working) {
      Store.show("running");
    } else {
      done();
    }
    return true;
  }

  function typed(event) {
    if (event.key === Qt.Key_Backspace && searching)
      Store.search = Store.search.slice(0, -1);
    else if (Keyboard.joins(Store.search, event.text) && !watching)
      Store.search += event.text.toLowerCase();
    else
      return false;
    zone = "rows";
    first();
    return true;
  }

  function askKey(event) {
    if (event.key === Qt.Key_Return || event.key === Qt.Key_Enter)
      return answer(true);
    if (event.key === Qt.Key_Escape)
      return answer(false);
    return true;
  }

  function key(event) {
    if (asking)
      return askKey(event);
    if (event.key === Qt.Key_Escape)
      return leave();
    if (event.key === Qt.Key_Tab || event.key === Qt.Key_Backtab) {
      zone = zone === "pages" ? "rows" : "pages";
      return true;
    }
    if (zone === "pages")
      return pagesKey(event);
    const keys = {
      [Qt.Key_Up]: () => move(-1),
      [Qt.Key_Down]: () => move(1),
      [Qt.Key_Return]: () => activate(),
      [Qt.Key_Enter]: () => activate(),
      [Qt.Key_Space]: () => searching ? typed(event) : markFocused(),
      [Qt.Key_Home]: () => first() || true,
      [Qt.Key_PageUp]: () => move(-10),
      [Qt.Key_PageDown]: () => move(10)
    };
    return keys[event.key] ? keys[event.key]() : typed(event);
  }

  function pagesKey(event) {
    if (event.key === Qt.Key_Up || event.key === Qt.Key_Down)
      return turn(event.key === Qt.Key_Up ? -1 : 1);
    if (event.key === Qt.Key_Return || event.key === Qt.Key_Enter || event.key === Qt.Key_Right) {
      zone = "rows";
      return true;
    }
    return typed(event);
  }

  function takeKeys() {
    if (adding)
      form.start();
    else if (zone === "rows" && watching && Store.job)
      view.forceActiveFocus();
    else
      frame.forceActiveFocus();
  }

  focus: true
  onItemsChanged: {
    if (!selectable(items[index]))
      first();
    seen();
  }
  onZoneChanged: Qt.callLater(takeKeys)
  onAddingChanged: Qt.callLater(takeKeys)
  onWatchingChanged: Qt.callLater(takeKeys)
  Keys.onPressed: event => {
    Keyboard.used = true;
    if (frame.key(event))
      event.accepted = true;
  }

  Connections {
    target: Store

    function onPageChanged() {
      frame.adding = false;
      frame.first();
      Qt.callLater(frame.takeKeys);
    }

    function onJobChanged() {
      Qt.callLater(frame.takeKeys);
    }
  }

  Rectangle {
    id: sidebar

    anchors.left: parent.left
    anchors.top: parent.top
    anchors.bottom: parent.bottom
    width: Theme.size(240)
    color: Theme.surface

    Rectangle {
      id: search

      anchors.left: parent.left
      anchors.right: parent.right
      anchors.top: parent.top
      anchors.margins: 16
      height: Theme.size(32)
      radius: Theme.radius
      color: Theme.base
      border.width: frame.searching ? 2 : 1
      border.color: frame.searching ? Theme.accent : Theme.overlay

      Label {
        anchors.fill: parent
        anchors.leftMargin: 10
        anchors.rightMargin: 10
        text: frame.searching ? "  " + Store.search : "  Type to search"
        color: frame.searching ? Theme.text : Theme.subtext
        elide: Text.ElideLeft
      }
    }

    Column {
      anchors.left: parent.left
      anchors.right: parent.right
      anchors.top: search.bottom
      anchors.margins: 10
      anchors.topMargin: 12
      spacing: 2

      Repeater {
        model: Store.pages

        Choice {
          required property var modelData

          width: parent.width
          icon: modelData.glyph
          label: modelData.label
          detail: modelData.id === "running" && Store.working ? "Working" : modelData.id === "apps" && Store.marked > 0 ? Store.marked + " marked" : ""
          selected: modelData.id === Store.page && !frame.searching
          border.width: selected && frame.zone === "pages" ? 2 : 0
          border.color: Theme.accent
          onClicked: {
            Store.search = "";
            Store.show(modelData.id);
            frame.zone = "rows";
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

    anchors.left: rule.right
    anchors.right: parent.right
    anchors.top: parent.top
    anchors.bottom: strip.top

    Label {
      id: title

      anchors.left: parent.left
      anchors.right: parent.right
      anchors.top: parent.top
      anchors.leftMargin: 48
      anchors.rightMargin: 48
      anchors.topMargin: 24
      text: frame.searching ? "Software matching " + Store.search : Store.pageNow.label
      font.family: Theme.display
      font.pixelSize: Theme.size(22)
      font.bold: Theme.displayBold
      elide: Text.ElideRight
    }

    Label {
      id: line

      anchors.left: title.left
      anchors.right: title.right
      anchors.top: title.bottom
      anchors.topMargin: 4
      text: frame.searching ? "Apps first, then Debian packages, at most " + Store.most + " of them" : Store.pageNow.help
      color: Theme.subtext
      wrapMode: Text.Wrap
      maximumLineCount: 2
      elide: Text.ElideRight
    }

    ListView {
      id: list

      anchors.left: parent.left
      anchors.right: parent.right
      anchors.top: line.bottom
      anchors.bottom: said.visible ? said.top : parent.bottom
      anchors.leftMargin: 32
      anchors.rightMargin: 32
      anchors.topMargin: 16
      anchors.bottomMargin: 8
      visible: !frame.watching && !frame.adding
      clip: true
      spacing: 4
      model: frame.items
      currentIndex: frame.index
      highlightMoveDuration: 0
      boundsBehavior: Flickable.StopAtBounds
      onCurrentIndexChanged: positionViewAtIndex(currentIndex, ListView.Contain)

      delegate: Loader {
        id: holder

        required property var modelData
        required property int index

        width: ListView.view.width
        sourceComponent: modelData.kind === "section" ? sectionPart : rowPart

        Component {
          id: sectionPart

          Item {
            implicitHeight: heading.implicitHeight + 16

            Heading {
              id: heading

              anchors.left: parent.left
              anchors.right: parent.right
              anchors.bottom: parent.bottom
              anchors.leftMargin: 16
              text: holder.modelData.label
            }
          }
        }

        Component {
          id: rowPart

          AppRow {
            item: holder.modelData
            current: holder.index === frame.index && frame.zone === "rows"
            marked: ["app", "pkg"].includes(holder.modelData.kind) && Store.isMarked(holder.modelData.kind, holder.modelData.id)
            onPicked: {
              frame.zone = "rows";
              frame.index = holder.index;
              frame.activate();
            }
            onToggled: {
              frame.index = holder.index;
              frame.markFocused();
            }
          }
        }
      }
    }

    WebForm {
      id: form

      anchors.left: parent.left
      anchors.right: parent.right
      anchors.top: line.bottom
      anchors.leftMargin: 48
      anchors.rightMargin: 48
      anchors.topMargin: 24
      visible: frame.adding
      onAdd: (name, address, icon) => {
        Store.addWebApp(name, address, icon);
        frame.adding = false;
      }
      onCancel: frame.adding = false
    }

    Running {
      id: view

      anchors.left: parent.left
      anchors.right: parent.right
      anchors.top: line.bottom
      anchors.bottom: parent.bottom
      anchors.leftMargin: 32
      anchors.rightMargin: 32
      anchors.topMargin: 16
      visible: frame.watching
      onClosed: {
        Store.show("apps");
        frame.zone = "rows";
      }
    }

    Label {
      id: said

      anchors.left: title.left
      anchors.right: title.right
      anchors.bottom: parent.bottom
      anchors.bottomMargin: 12
      visible: Store.page === "webapps" && Store.said !== "" && !frame.searching
      text: (Store.failed ? "▨ FAIL  " : "■ PASS  ") + Store.said
      elide: Text.ElideRight
    }
  }

  Rectangle {
    id: strip

    anchors.left: rule.right
    anchors.right: parent.right
    anchors.bottom: parent.bottom
    height: bar.implicitHeight + 24
    color: Theme.surface

    Row {
      id: bar

      anchors.left: parent.left
      anchors.right: parent.right
      anchors.verticalCenter: parent.verticalCenter
      anchors.leftMargin: 32
      anchors.rightMargin: 32
      spacing: 16

      Label {
        id: asked

        anchors.verticalCenter: parent.verticalCenter
        visible: !!frame.asking
        width: Math.min(implicitWidth, bar.width - yes.width - no.width - 48)
        text: frame.asking ? frame.asking.text : ""
        font.bold: true
        elide: Text.ElideRight
      }

      Button {
        id: yes

        visible: !!frame.asking
        label: "Yes"
        primary: true
        onClicked: frame.answer(true)
      }

      Button {
        id: no

        visible: !!frame.asking
        label: "No"
        onClicked: frame.answer(false)
      }

      Button {
        visible: !frame.asking && Store.marked > 0
        label: "Apply " + Store.marked + (Store.marked === 1 ? " change" : " changes")
        primary: true
        onClicked: frame.askChanges(Store.changes())
      }

      Hints {
        anchors.verticalCenter: parent.verticalCenter
        visible: !frame.asking
        pairs: frame.hints
      }
    }
  }
}
