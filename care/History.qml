import QtQuick
import QtQuick.Layouts
import Quickshell
import qs.common
import "." as Care

FocusScope {
  id: page

  property string zone: "points"
  property int pointAt: 0
  property int fileAt: 0
  property string folder: Store.homeDir
  readonly property var home: Store.home
  readonly property var backup: Store.backup
  readonly property bool backingUp: !!backup && !backup.missing && backup.location !== ""
  readonly property var points: timeline(home, Store.backups)
  readonly property var point: points.length ? points[Math.min(pointAt, points.length - 1)] : null
  readonly property var entries: Store.files || []
  readonly property var entry: entries.length ? entries[Math.min(fileAt, entries.length - 1)] : null
  readonly property string shownFolder: folder === Store.homeDir ? "~" : "~/" + folder.slice(Store.homeDir.length + 1)
  readonly property var actions: [now, daily, password, settings].filter(item => item.visible)
  readonly property var hints: zone === "actions" ? [[["Left", "Right"], "pick"], ["Enter", "press"], ["Down", "points"]] : zone === "files" ? [[["Up", "Down"], "file"], ["Enter", "open"], ["Left", "back"], ["R", "restore"], ["C", "restore a copy"]] : [[["Up", "Down"], "point"], ["Enter", "its files"]]

  function size(bytes) {
    const units = ["B", "KiB", "MiB", "GiB", "TiB"];
    let at = 0;
    let value = bytes;
    while (value >= 1024 && at < units.length - 1) {
      value /= 1024;
      at++;
    }
    return (at ? value.toFixed(1) : String(value)) + " " + units[at];
  }

  function when(time) {
    return Qt.formatDateTime(new Date(time), "ddd d MMM yyyy, HH:mm");
  }

  function timeline() {
    const found = (home && home.points ? home.points : []).map(point => ({
          kind: "home",
          id: point.number,
          time: point.time,
          label: "Snapshot of /home"
        }));
    for (const item of Store.backups || [])
      found.push({
        kind: "backup",
        id: item.id,
        time: item.time,
        label: "Backup"
      });
    return found.sort((first, second) => new Date(second.time) - new Date(first.time));
  }

  function homeWords() {
    if (!home)
      return "Looking for the snapshots of /home";
    if (home.state === "none")
      return "There are none: /home isn't a btrfs subvolume of its own here.";
    const count = home.points.length === 1 ? "1 is kept now." : home.points.length + " are kept now.";
    if (home.state === "off")
      return "Off: no new ones are taken. " + count + " Turn them on in Settings, under Updates.";
    return "One every hour; the last " + home.hourly + " hours and " + home.daily + " days are kept, with ~/.cache and Steam's games left out. " + count;
  }

  function backupWords() {
    if (!backup)
      return "Looking for the backups";
    if (backup.missing)
      return "restic-backup isn't set up: the Backups part of the setup installs it.";
    if (!backup.location)
      return "No backup location yet. Choose one in a terminal: restic-backup init followed by a folder, a USB drive, or s3:, b2: or rclone:.";
    const last = backup.last ? "The last backup: " + when(backup.last) + "." : "No backup yet.";
    const plan = backup.schedule === "daily" ? " Once a day, on the charger." : " Only when you start one" + (backup.usb ? " or plug in the drive." : ".");
    return "To " + backup.location + ". " + last + plan;
  }

  function enter() {
    page.forceActiveFocus();
  }

  function look() {
    if (point)
      Store.browse(point, folder);
  }

  function pickPoint(at) {
    pointAt = Math.max(0, Math.min(points.length - 1, at));
    fileAt = 0;
    look();
  }

  function open() {
    if (!entry || entry.kind !== "dir")
      return;
    folder = folder + "/" + entry.name;
    fileAt = 0;
    look();
  }

  function up() {
    if (folder === Store.homeDir) {
      zone = "points";
      return;
    }
    folder = folder.slice(0, folder.lastIndexOf("/"));
    fileAt = 0;
    look();
  }

  function bring(copy) {
    if (point && entry && !Store.historyBusy)
      Store.bringBack(point, folder + "/" + entry.name, copy);
  }

  function sideways(step) {
    const at = actions.findIndex(item => item.activeFocus);
    actions[Math.max(0, Math.min(actions.length - 1, at + step))].forceActiveFocus();
    return true;
  }

  function actionKey(key) {
    if (key === Qt.Key_Left || key === Qt.Key_Right)
      return sideways(key === Qt.Key_Left ? -1 : 1);
    if (key !== Qt.Key_Down)
      return false;
    zone = "points";
    page.forceActiveFocus();
    return true;
  }

  function pointKey(key) {
    if (key === Qt.Key_Up && pointAt === 0 && actions.length) {
      zone = "actions";
      actions[0].forceActiveFocus();
    } else if (key === Qt.Key_Up || key === Qt.Key_Down) {
      pickPoint(pointAt + (key === Qt.Key_Up ? -1 : 1));
    } else if ([Qt.Key_Right, Qt.Key_Return, Qt.Key_Enter].includes(key) && point) {
      zone = "files";
      look();
    } else {
      return false;
    }
    return true;
  }

  function fileKey(key) {
    const moves = new Map([[Qt.Key_Up, -1], [Qt.Key_Down, 1], [Qt.Key_PageUp, -10], [Qt.Key_PageDown, 10]]);
    if (moves.has(key))
      fileAt = Math.max(0, Math.min(entries.length - 1, fileAt + moves.get(key)));
    else if (key === Qt.Key_Return || key === Qt.Key_Enter || key === Qt.Key_Right)
      open();
    else if (key === Qt.Key_Left || key === Qt.Key_Backspace)
      up();
    else if (key === Qt.Key_R || key === Qt.Key_C)
      bring(key === Qt.Key_C);
    else
      return false;
    return true;
  }

  function key(event) {
    if (zone === "actions")
      return actionKey(event.key);
    return zone === "files" ? fileKey(event.key) : pointKey(event.key);
  }

  Keys.onPressed: event => event.accepted = page.key(event)
  onPointChanged: Qt.callLater(look)
  Component.onCompleted: Store.loadHistory()

  ColumnLayout {
    anchors.fill: parent
    spacing: 16

    Heading {
      Layout.fillWidth: true
      text: "Snapshots of /home"
      note: page.home ? ({
          on: "on",
          off: "off",
          none: "none"
        })[page.home.state] || "" : ""
    }

    Label {
      Layout.fillWidth: true
      text: page.homeWords()
      wrapMode: Text.Wrap
    }

    Heading {
      Layout.fillWidth: true
      Layout.topMargin: 8
      text: "Backups"
      note: page.backingUp ? (page.backup.schedule === "daily" ? "daily" : "by hand") : ""
    }

    Label {
      Layout.fillWidth: true
      text: page.backupWords()
      wrapMode: Text.Wrap
    }

    RowLayout {
      Layout.fillWidth: true
      spacing: 8
      visible: page.backingUp && page.backup.stale

      Care.Mark {
        kind: "warn"
      }

      Label {
        Layout.fillWidth: true
        text: "WARN  No backup for a week. Plug in the charger and the backup drive, or back up now."
        wrapMode: Text.Wrap
      }
    }

    RowLayout {
      Layout.fillWidth: true
      spacing: 16

      Button {
        id: now

        visible: page.backingUp && !Store.historyBusy
        label: "Back up now"
        primary: true
        onClicked: Store.act("backup", ["env", Store.backupTool, "run"])
      }

      Button {
        id: daily

        visible: page.backingUp && !Store.historyBusy
        label: page.backingUp && page.backup.schedule === "daily" ? "Stop daily backups" : "Back up daily"
        onClicked: Store.act("schedule", ["env", Store.backupTool, "schedule", page.backup.schedule === "daily" ? "off" : "daily"])
      }

      Button {
        id: password

        visible: page.backingUp && Store.secret === ""
        label: "Show the password"
        onClicked: Store.showSecret()
      }

      Button {
        id: settings

        visible: !!page.home && page.home.state !== "none"
        label: "Settings"
        onClicked: Quickshell.execDetached(["solanum", "settings", "updates"])
      }

      Item {
        Layout.fillWidth: true
      }
    }

    Rectangle {
      Layout.fillWidth: true
      Layout.preferredHeight: shownSecret.implicitHeight + 32
      visible: Store.secret !== "" || Store.secretWhy !== ""
      color: Theme.surface
      radius: Theme.radius
      border.width: 1
      border.color: Theme.overlay

      RowLayout {
        id: shownSecret

        anchors.fill: parent
        anchors.margins: 16
        spacing: 16

        Label {
          Layout.fillWidth: true
          text: Store.secret !== "" ? "The backup password: " + Store.secret + "\nKeep a copy somewhere safe: without it the backups can't be read." : Store.secretWhy
          wrapMode: Text.WrapAnywhere
        }

        Button {
          visible: Store.secret !== ""
          label: "Hide"
          onClicked: Store.secret = ""
        }
      }
    }

    Heading {
      Layout.fillWidth: true
      Layout.topMargin: 8
      text: "Timeline"
      note: page.point ? page.shownFolder : ""
    }

    RowLayout {
      Layout.fillWidth: true
      Layout.fillHeight: true
      spacing: 16

      ListView {
        id: pointList

        Layout.preferredWidth: Math.round(parent.width * 0.38)
        Layout.fillHeight: true
        clip: true
        spacing: 4
        model: page.points
        currentIndex: page.pointAt
        highlightMoveDuration: 0
        boundsBehavior: Flickable.StopAtBounds
        onCurrentIndexChanged: positionViewAtIndex(currentIndex, ListView.Contain)

        delegate: Care.ListRow {
          required property var modelData
          required property int index

          width: ListView.view.width
          title: page.when(modelData.time)
          detail: modelData.label
          current: index === page.pointAt
          framed: page.activeFocus && page.zone === "points"
          onPicked: {
            page.zone = "points";
            page.pickPoint(index);
            page.forceActiveFocus();
          }
        }
      }

      ColumnLayout {
        Layout.fillWidth: true
        Layout.preferredWidth: 1
        Layout.fillHeight: true
        spacing: 8

        Label {
          Layout.fillWidth: true
          visible: page.points.length === 0
          text: Store.backupsWhy !== "" && !Store.backups ? "The backups can't be read: " + Store.backupsWhy : "Nothing to restore from yet."
          color: Theme.subtext
          wrapMode: Text.Wrap
        }

        Label {
          Layout.fillWidth: true
          visible: !!page.point && (Store.files === null || Store.filesWhy !== "" && page.entries.length === 0)
          text: Store.files === null ? "RUNNING  Reading " + page.shownFolder : Store.filesWhy
          color: Theme.subtext
          wrapMode: Text.Wrap
        }

        ListView {
          id: fileList

          Layout.fillWidth: true
          Layout.fillHeight: true
          clip: true
          spacing: 2
          model: page.point ? page.entries : []
          currentIndex: page.fileAt
          highlightMoveDuration: 0
          boundsBehavior: Flickable.StopAtBounds
          onCurrentIndexChanged: positionViewAtIndex(currentIndex, ListView.Contain)

          delegate: Care.ListRow {
            required property var modelData
            required property int index

            width: ListView.view.width
            title: modelData.name + (modelData.kind === "dir" ? "/" : "")
            detail: modelData.kind === "dir" ? "" : page.size(modelData.size) + ", " + page.when(modelData.time)
            current: index === page.fileAt && page.zone === "files"
            framed: page.activeFocus
            onPicked: {
              page.zone = "files";
              page.fileAt = index;
              page.forceActiveFocus();
            }
          }
        }

        RowLayout {
          Layout.fillWidth: true
          spacing: 16
          visible: !!page.entry && page.zone === "files"

          Button {
            visible: !Store.historyBusy
            label: "Restore"
            primary: true
            onClicked: page.bring(false)
          }

          Button {
            visible: !Store.historyBusy
            label: "Restore a copy"
            onClicked: page.bring(true)
          }

          Item {
            Layout.fillWidth: true
          }
        }
      }
    }

    RowLayout {
      Layout.fillWidth: true
      spacing: 8
      visible: Store.historyBusy || Store.said !== ""

      Care.Mark {
        Layout.alignment: Qt.AlignTop
        kind: Store.historyBusy ? "busy" : Store.saidCode === 0 ? "ok" : "fail"
      }

      Label {
        Layout.fillWidth: true
        text: Store.historyBusy ? "RUNNING  " + ({
            backup: "Backing up",
            schedule: "Changing the schedule",
            restore: "Restoring",
            copy: "Restoring a copy"
          })[Store.doing] : (Store.saidCode === 0 ? "PASS  " : "FAIL  ") + Store.said
        wrapMode: Text.WrapAnywhere
      }
    }
  }
}
