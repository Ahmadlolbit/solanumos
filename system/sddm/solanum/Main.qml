import QtQuick
import QtQuick.Controls
import QtQuick.Effects

Rectangle {
  id: root

  readonly property bool contrast: chosen.contrast
  readonly property real textScale: chosen.larger ? 1.5 : 1
  readonly property real keyScale: Math.max(0.5, Math.min(textScale, (height - size(320)) / 276))
  readonly property color background: contrast ? "#000000" : config.background || "#1e1e2e"
  readonly property color foreground: contrast ? "#ffffff" : config.foreground || "#cdd6f4"
  readonly property color accent: contrast ? "#ffff00" : config.accent || "#89b4fa"
  readonly property color fail: contrast ? "#ff6e6e" : config.fail || "#f38ba8"
  readonly property color muted: contrast ? root.foreground : Qt.alpha(root.foreground, 0.6)
  readonly property string fontFamily: config.font || "monospace"
  readonly property string displayFamily: Qt.fontFamilies().includes("Departure Mono") ? "Departure Mono" : root.fontFamily
  readonly property string timeFormat: config.clock === "12h" ? "h:mm AP" : "HH:mm"
  property int userIndex: userModel.lastIndex >= 0 ? userModel.lastIndex : 0
  readonly property var currentUser: users.count > 0 ? users.itemAt(Math.min(userIndex, users.count - 1)) : null
  readonly property string userName: currentUser ? currentUser.login : userModel.lastUser
  readonly property int sessionIndex: sessionModel.lastIndex >= 0 ? sessionModel.lastIndex : 0
  readonly property var layout: keyboard.layouts && keyboard.currentLayout >= 0 && keyboard.currentLayout < keyboard.layouts.length ? keyboard.layouts[keyboard.currentLayout] : null
  readonly property string layoutName: layout ? layout.longName || layout.shortName : ""
  property bool failed: false
  property int tries: 0
  property date now: new Date()

  width: 1920
  height: 1080
  color: root.background

  function size(pixels) {
    return Math.round(pixels * root.textScale);
  }

  function hint() {
    if (access.focused)
      return "Enter to turn it on or off · Up and Down to move · Esc to close";
    if (button.activeFocus)
      return "Enter for larger text, high contrast or a keyboard · Tab to go on";
    if (!password.activeFocus)
      return "Enter to press · Left and Right to move · Tab to go back";
    return "Enter to log in" + (users.count > 1 ? " · F1 for the next account" : "") + " · Tab for accessibility, sleep, restart or shut down";
  }

  function login() {
    failed = false;
    sddm.login(root.userName, password.text, root.sessionIndex);
  }

  function nextUser() {
    if (users.count < 2)
      return;
    userIndex = (Math.min(userIndex, users.count - 1) + 1) % users.count;
    failed = false;
    tries = 0;
    password.text = "";
    password.forceActiveFocus();
  }

  QtObject {
    id: chosen

    property bool larger: false
    property bool contrast: false
    property bool keyboard: false

    objectName: "chosen"
  }

  Loader {
    id: memory

    objectName: "memory"
    source: "Remembered.qml"
    onLoaded: {
      chosen.larger = item.larger;
      chosen.contrast = item.contrast;
      chosen.keyboard = item.keyboard;
    }
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
      root.tries += 1;
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
    visible: !root.contrast
    source: config.wallpaper ? Qt.resolvedUrl(config.wallpaper) : ""
    fillMode: Image.PreserveAspectCrop
    asynchronous: true
  }

  MultiEffect {
    anchors.fill: parent
    source: wallpaper
    visible: !root.contrast && wallpaper.status === Image.Ready
    blurEnabled: true
    blur: 1.0
    blurMax: 48
  }

  Rectangle {
    anchors.fill: parent
    visible: !root.contrast && wallpaper.status === Image.Ready
    color: "black"
    opacity: 0.4
  }

  Column {
    objectName: "middle"
    anchors.horizontalCenter: parent.horizontalCenter
    y: keys.visible ? Math.max(root.size(16), keys.y - height - root.size(24)) : (parent.height - height) / 2 - root.size(40)
    spacing: root.size(14)

    Text {
      anchors.horizontalCenter: parent.horizontalCenter
      visible: !keys.visible
      text: Qt.formatTime(root.now, root.timeFormat)
      color: root.foreground
      font.family: root.displayFamily
      font.pixelSize: root.size(132)
    }

    Text {
      anchors.horizontalCenter: parent.horizontalCenter
      visible: !keys.visible
      text: Qt.formatDate(root.now, "dddd, d MMMM")
      color: root.foreground
      font.family: root.fontFamily
      font.pixelSize: root.size(26)
    }

    Item {
      width: 1
      height: root.size(56)
      visible: !keys.visible
    }

    Text {
      anchors.horizontalCenter: parent.horizontalCenter
      text: root.currentUser ? root.currentUser.shown : root.userName
      color: Qt.alpha(root.foreground, 0.8)
      font.family: root.fontFamily
      font.pixelSize: root.size(18)

      Text {
        anchors.left: parent.right
        anchors.leftMargin: 16
        anchors.baseline: parent.baseline
        visible: users.count > 1
        text: Math.min(root.userIndex, users.count - 1) + 1 + " of " + users.count
        color: root.muted
        font.family: root.fontFamily
        font.pixelSize: root.size(16)
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
      width: root.size(320)
      height: root.size(56)
      color: Qt.alpha(root.background, 0.8)
      border.color: root.failed ? root.fail : password.activeFocus ? root.accent : root.muted
      border.width: 2

      TextField {
        id: password

        objectName: "password"
        anchors.fill: parent
        anchors.margins: 2
        focus: true
        echoMode: TextInput.Password
        passwordCharacter: "•"
        horizontalAlignment: TextInput.AlignHCenter
        cursorVisible: activeFocus && text !== ""
        color: root.foreground
        font.family: root.fontFamily
        font.pixelSize: root.size(20)
        background: Item {}
        onTextEdited: root.failed = false
        onAccepted: root.login()
        Keys.onEscapePressed: text = ""
        Keys.onTabPressed: button.forceActiveFocus()
        Keys.onBacktabPressed: power.step(actions.count, -1)
      }

      Text {
        anchors.centerIn: parent
        visible: password.text === ""
        text: "Password"
        color: root.foreground
        opacity: root.contrast ? 1 : 0.5
        font.family: root.fontFamily
        font.pixelSize: root.size(20)
      }
    }

    Text {
      anchors.horizontalCenter: parent.horizontalCenter
      text: root.layoutName ? "\u{f030c} " + root.layoutName : " "
      color: root.muted
      font.family: root.fontFamily
      font.pixelSize: root.size(16)
    }

    Text {
      anchors.horizontalCenter: parent.horizontalCenter
      text: [root.failed ? "Wrong password · try " + root.tries : "", keyboard.capsLock ? "Caps Lock is on" : ""].filter(Boolean).join(" · ") || " "
      color: root.failed ? root.fail : root.foreground
      font.family: root.fontFamily
      font.pixelSize: root.size(16)
    }
  }

  Text {
    anchors.horizontalCenter: parent.horizontalCenter
    anchors.bottom: parent.bottom
    anchors.bottomMargin: 32
    text: root.hint()
    color: root.muted
    font.family: root.fontFamily
    font.pixelSize: root.size(14)
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
        width: root.size(96)
        height: root.size(72)
        color: "transparent"
        border.color: root.accent
        border.width: activeFocus ? 2 : 0
        Keys.onLeftPressed: power.step(index, -1)
        Keys.onRightPressed: power.step(index, 1)
        Keys.onReturnPressed: press()
        Keys.onEnterPressed: press()
        Keys.onSpacePressed: press()
        Keys.onTabPressed: password.forceActiveFocus()
        Keys.onBacktabPressed: button.forceActiveFocus()
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
          font.pixelSize: root.size(26)
        }

        Text {
          anchors.horizontalCenter: parent.horizontalCenter
          anchors.bottom: parent.bottom
          anchors.bottomMargin: 8
          visible: action.lit
          text: action.modelData[3]
          color: root.foreground
          font.family: root.fontFamily
          font.pixelSize: root.size(14)
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

  Keyboard {
    id: keys

    objectName: "keys"
    anchors.horizontalCenter: parent.horizontalCenter
    anchors.bottom: parent.bottom
    anchors.bottomMargin: root.size(120)
    visible: chosen.keyboard
    target: password
    ink: root.foreground
    paper: root.background
    accent: root.accent
    family: root.fontFamily
    factor: root.keyScale
    onTyped: root.failed = false
    onEntered: root.login()
  }

  Column {
    id: access

    property bool open: false
    readonly property bool focused: panel.activeFocus

    anchors.left: parent.left
    anchors.bottom: parent.bottom
    anchors.margins: 32
    spacing: 8

    function show(at) {
      const shown = Math.max(0, Math.min(at, choices.count - 1));
      choices.itemAt(shown).forceActiveFocus();
    }

    function close() {
      open = false;
      button.forceActiveFocus();
    }

    FocusScope {
      id: panel

      visible: access.open
      width: list.width
      height: list.height

      Column {
        id: list

        spacing: 8

        Repeater {
          id: choices

          model: [["larger", "Larger text"], ["contrast", "High contrast"], ["keyboard", "Keyboard"]]

          Rectangle {
            id: choice

            required property var modelData
            required property int index
            readonly property bool on: chosen[modelData[0]]

            objectName: modelData[0]
            width: root.size(260)
            height: root.size(44)
            color: on ? root.foreground : Qt.alpha(root.background, 0.8)
            border.color: root.accent
            border.width: activeFocus ? 2 : 0
            Keys.onUpPressed: access.show(index - 1)
            Keys.onDownPressed: index === choices.count - 1 ? button.forceActiveFocus() : access.show(index + 1)
            Keys.onReturnPressed: press()
            Keys.onEnterPressed: press()
            Keys.onSpacePressed: press()
            Keys.onEscapePressed: access.close()
            Keys.onTabPressed: power.step(-1, 1)
            Keys.onBacktabPressed: password.forceActiveFocus()

            function press() {
              const name = modelData[0];
              chosen[name] = !chosen[name];
              if (!memory.item)
                return;
              memory.item[name] = chosen[name];
              memory.item.setValue(name, chosen[name]);
              memory.item.sync();
            }

            Text {
              anchors.left: parent.left
              anchors.leftMargin: 16
              anchors.verticalCenter: parent.verticalCenter
              text: choice.modelData[1]
              color: choice.on ? root.background : root.foreground
              font.family: root.fontFamily
              font.pixelSize: root.size(16)
            }

            Text {
              anchors.right: parent.right
              anchors.rightMargin: 16
              anchors.verticalCenter: parent.verticalCenter
              text: choice.on ? "On" : "Off"
              color: choice.on ? root.background : root.foreground
              font.family: root.fontFamily
              font.pixelSize: root.size(14)
            }

            MouseArea {
              anchors.fill: parent
              hoverEnabled: true
              cursorShape: Qt.PointingHandCursor
              onClicked: choice.press()
            }
          }
        }

        Text {
          objectName: "untilBoot"
          visible: !memory.item
          text: "Only until the next boot"
          color: root.muted
          font.family: root.fontFamily
          font.pixelSize: root.size(14)
        }
      }
    }

    Rectangle {
      id: button

      readonly property bool lit: activeFocus || hovered.containsMouse

      objectName: "access"
      width: root.size(136)
      height: root.size(72)
      color: "transparent"
      border.color: root.accent
      border.width: activeFocus ? 2 : 0
      Keys.onReturnPressed: press()
      Keys.onEnterPressed: press()
      Keys.onSpacePressed: press()
      Keys.onUpPressed: access.open ? access.show(choices.count - 1) : press()
      Keys.onTabPressed: power.step(-1, 1)
      Keys.onBacktabPressed: password.forceActiveFocus()
      Keys.onEscapePressed: access.open ? access.close() : password.forceActiveFocus()

      function press() {
        access.open = !access.open;
        if (access.open)
          access.show(0);
      }

      Text {
        anchors.horizontalCenter: parent.horizontalCenter
        anchors.top: parent.top
        anchors.topMargin: 8
        text: "\uf29a"
        color: button.lit || access.open ? root.accent : root.foreground
        font.family: root.fontFamily
        font.pixelSize: root.size(26)
      }

      Text {
        anchors.horizontalCenter: parent.horizontalCenter
        anchors.bottom: parent.bottom
        anchors.bottomMargin: 8
        visible: button.lit || access.open
        text: "Accessibility"
        color: root.foreground
        font.family: root.fontFamily
        font.pixelSize: root.size(14)
      }

      MouseArea {
        id: hovered

        anchors.fill: parent
        hoverEnabled: true
        cursorShape: Qt.PointingHandCursor
        onClicked: button.press()
      }
    }
  }
}
