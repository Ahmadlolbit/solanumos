import QtQuick
import QtQuick.Controls
import QtQuick.Effects

Rectangle {
  id: root

  readonly property color background: config.background || "#1e1e2e"
  readonly property color foreground: config.foreground || "#cdd6f4"
  readonly property color accent: config.accent || "#89b4fa"
  readonly property color fail: config.fail || "#f38ba8"
  readonly property color muted: Qt.alpha(root.foreground, 0.6)
  readonly property string fontFamily: config.font || "monospace"
  readonly property string timeFormat: config.clock === "12h" ? "h:mm AP" : "HH:mm"
  property int userIndex: userModel.lastIndex >= 0 ? userModel.lastIndex : 0
  readonly property var currentUser: users.count > 0 ? users.itemAt(Math.min(userIndex, users.count - 1)) : null
  readonly property string userName: currentUser ? currentUser.login : userModel.lastUser
  readonly property int sessionIndex: sessionModel.lastIndex >= 0 ? sessionModel.lastIndex : 0
  property bool failed: false
  property date now: new Date()

  width: 1920
  height: 1080
  color: root.background

  function login() {
    failed = false;
    sddm.login(root.userName, password.text, root.sessionIndex);
  }

  function nextUser() {
    if (users.count < 2)
      return;
    userIndex = (Math.min(userIndex, users.count - 1) + 1) % users.count;
    failed = false;
    password.text = "";
    password.forceActiveFocus();
  }

  Shortcut {
    sequences: ["F1", "Alt+U"]
    enabled: users.count > 1
    onActivated: root.nextUser()
  }

  Connections {
    target: sddm

    function onLoginFailed() {
      root.failed = true;
      password.text = "";
      password.forceActiveFocus();
    }
  }

  Timer {
    interval: 1000
    running: true
    repeat: true
    onTriggered: root.now = new Date()
  }

  Repeater {
    id: users

    model: userModel

    delegate: Item {
      required property string name
      required property string realName
      readonly property string login: name
      readonly property string shown: realName || name
    }
  }

  Image {
    id: wallpaper

    anchors.fill: parent
    source: config.wallpaper ? Qt.resolvedUrl(config.wallpaper) : ""
    fillMode: Image.PreserveAspectCrop
    asynchronous: true
  }

  MultiEffect {
    anchors.fill: parent
    source: wallpaper
    visible: wallpaper.status === Image.Ready
    blurEnabled: true
    blur: 1.0
    blurMax: 48
  }

  Rectangle {
    anchors.fill: parent
    visible: wallpaper.status === Image.Ready
    color: "black"
    opacity: 0.4
  }

  Column {
    anchors.centerIn: parent
    anchors.verticalCenterOffset: -40
    spacing: 14

    Text {
      anchors.horizontalCenter: parent.horizontalCenter
      text: Qt.formatTime(root.now, root.timeFormat)
      color: root.foreground
      font.family: root.fontFamily
      font.pixelSize: 128
    }

    Text {
      anchors.horizontalCenter: parent.horizontalCenter
      text: Qt.formatDate(root.now, "dddd, d MMMM")
      color: root.foreground
      font.family: root.fontFamily
      font.pixelSize: 26
    }

    Item {
      width: 1
      height: 56
    }

    Text {
      anchors.horizontalCenter: parent.horizontalCenter
      text: root.currentUser ? root.currentUser.shown : root.userName
      color: Qt.alpha(root.foreground, 0.8)
      font.family: root.fontFamily
      font.pixelSize: 18

      Text {
        anchors.left: parent.right
        anchors.leftMargin: 16
        anchors.baseline: parent.baseline
        visible: users.count > 1
        text: Math.min(root.userIndex, users.count - 1) + 1 + " of " + users.count
        color: root.muted
        font.family: root.fontFamily
        font.pixelSize: 16
      }

      MouseArea {
        anchors.fill: parent
        enabled: users.count > 1
        cursorShape: Qt.PointingHandCursor
        onClicked: root.nextUser()
      }
    }

    Rectangle {
      anchors.horizontalCenter: parent.horizontalCenter
      width: 320
      height: 56
      color: Qt.alpha(root.background, 0.8)
      border.color: root.failed ? root.fail : password.activeFocus ? root.accent : root.muted
      border.width: 2

      TextField {
        id: password

        anchors.fill: parent
        anchors.margins: 2
        focus: true
        echoMode: TextInput.Password
        passwordCharacter: "•"
        horizontalAlignment: TextInput.AlignHCenter
        cursorVisible: activeFocus && text !== ""
        color: root.foreground
        font.family: root.fontFamily
        font.pixelSize: 20
        background: Item {}
        onTextEdited: root.failed = false
        onAccepted: root.login()
        Keys.onEscapePressed: text = ""
        Keys.onTabPressed: power.step(-1, 1)
        Keys.onBacktabPressed: power.step(actions.count, -1)
      }

      Text {
        anchors.centerIn: parent
        visible: password.text === ""
        text: "Password"
        color: root.foreground
        opacity: 0.5
        font.family: root.fontFamily
        font.pixelSize: 20
      }
    }

    Text {
      anchors.horizontalCenter: parent.horizontalCenter
      text: root.failed ? "Wrong password" : keyboard.capsLock ? "Caps Lock is on" : " "
      color: root.failed ? root.fail : root.foreground
      font.family: root.fontFamily
      font.pixelSize: 16
    }
  }

  Text {
    anchors.horizontalCenter: parent.horizontalCenter
    anchors.bottom: parent.bottom
    anchors.bottomMargin: 32
    text: !password.activeFocus ? "Enter to press · Left and Right to move · Tab to go back" : users.count > 1 ? "Enter to log in · F1 for the next account · Tab to sleep, restart or shut down" : "Enter to log in · Tab to sleep, restart or shut down"
    color: root.muted
    font.family: root.fontFamily
    font.pixelSize: 14
  }

  Row {
    id: power

    anchors.right: parent.right
    anchors.bottom: parent.bottom
    anchors.margins: 32
    spacing: 8

    function step(from, delta) {
      for (let at = from + delta; at >= 0 && at < actions.count; at += delta) {
        if (actions.itemAt(at).visible) {
          actions.itemAt(at).forceActiveFocus();
          return;
        }
      }
    }

    Repeater {
      id: actions

      model: [["\u{f04b2}", sddm.canSuspend, "suspend", "Sleep"], ["\u{f0709}", sddm.canReboot, "reboot", "Restart"], ["\u{f0425}", sddm.canPowerOff, "powerOff", "Shut down"]]

      Rectangle {
        id: action

        required property var modelData
        required property int index
        readonly property bool lit: activeFocus || hover.containsMouse

        visible: modelData[1]
        width: 96
        height: 72
        color: "transparent"
        border.color: root.accent
        border.width: activeFocus ? 2 : 0
        Keys.onLeftPressed: power.step(index, -1)
        Keys.onRightPressed: power.step(index, 1)
        Keys.onReturnPressed: press()
        Keys.onEnterPressed: press()
        Keys.onSpacePressed: press()
        Keys.onTabPressed: password.forceActiveFocus()
        Keys.onBacktabPressed: password.forceActiveFocus()
        Keys.onEscapePressed: password.forceActiveFocus()

        function press() {
          sddm[modelData[2]]();
        }

        Text {
          anchors.horizontalCenter: parent.horizontalCenter
          anchors.top: parent.top
          anchors.topMargin: 8
          text: action.modelData[0]
          color: action.lit ? root.accent : root.foreground
          font.family: root.fontFamily
          font.pixelSize: 26
        }

        Text {
          anchors.horizontalCenter: parent.horizontalCenter
          anchors.bottom: parent.bottom
          anchors.bottomMargin: 8
          visible: action.lit
          text: action.modelData[3]
          color: root.foreground
          font.family: root.fontFamily
          font.pixelSize: 14
        }

        MouseArea {
          id: hover

          anchors.fill: parent
          hoverEnabled: true
          cursorShape: Qt.PointingHandCursor
          onClicked: action.press()
        }
      }
    }
  }
}
