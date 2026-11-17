import QtQuick
import QtQuick.Layouts

FocusScope {
  id: root

  property Task task: null
  property string title: ""
  property string stopQuestion: "Stop it?"
  property bool logShown: false
  property var halting: null
  property bool held: false
  property bool rowsFollow: true
  property bool logFollow: true
  property int phase: 0
  readonly property bool running: task !== null && task.running
  readonly property int code: task ? task.code : -1
  readonly property var question: halting || (task ? task.question : null)
  readonly property bool choosing: question !== null && question.kind === "choose"
  readonly property string condition: running ? (task.stopping ? "Stopping" : task.question ? "Waiting for an answer" : "Running") : code === -1 ? "" : ({
        0: "Done",
        126: "Cancelled",
        130: "Stopped"
      })[code] || "Failed"
  readonly property string reason: !running && code !== -1 && code !== 0 && code !== 130 ? failure() : ""
  readonly property var hints: (choosing ? [[["Left", "Right"], "pick"], ["Enter", "answer"]] : question ? [["Enter", "answer"]] : []).concat([["L", logShown ? "hide log" : "show log"], [["Up", "Down"], "scroll"], ["Tab", "next"]])
  readonly property int textColumn: 8 + Theme.size(12) + 8 + Math.ceil(widest.advanceWidth) + 8

  signal closed

  function failure() {
    const line = task.reason.split("\n").filter(text => text.trim()).pop() || "";
    const wanted = line.toLowerCase();
    for (let index = 0; index < task.rows.count; index++) {
      if (task.rows.get(index).text.toLowerCase().includes(wanted))
        return "";
    }
    return line;
  }

  function askStop(text, items, word) {
    if (running && !task.stopping && !held)
      halting = {
        kind: "choose",
        text: text,
        items: items,
        word: word
      };
  }

  function answered(text) {
    const asking = halting;
    halting = null;
    if (!asking)
      task.answer(text);
    else if (text === asking.word)
      task.stop();
  }

  function settle() {
    if (held)
      return;
    if (choosing)
      choices.pick(Math.max(0, question.items.indexOf("No")));
    else if (question)
      field.focusInput();
    else if (code !== -1 && !running)
      ending.forceActiveFocus();
    else
      idle.forceActiveFocus();
  }

  function resume() {
    if (activeFocus)
      settle();
  }

  function scrolled(view, key) {
    const line = Theme.size(20);
    const steps = new Map([[Qt.Key_Up, -line], [Qt.Key_Down, line], [Qt.Key_PageUp, line - view.height], [Qt.Key_PageDown, view.height - line]]);
    if (key === Qt.Key_Home)
      view.positionViewAtBeginning();
    else if (key === Qt.Key_End)
      view.positionViewAtEnd();
    else if (steps.has(key))
      view.contentY = Math.max(view.originY, Math.min(view.contentY + steps.get(key), view.originY + view.contentHeight - view.height));
    else
      return false;
    return true;
  }

  function pressed(event) {
    if (event.modifiers & (Qt.ControlModifier | Qt.AltModifier | Qt.MetaModifier))
      return false;
    if (halting && event.key === Qt.Key_Escape) {
      halting = null;
      return true;
    }
    if (event.key === Qt.Key_L) {
      logShown = !logShown;
      return true;
    }
    if (choosing && (event.key === Qt.Key_Left || event.key === Qt.Key_Right))
      return choices.step(event.key === Qt.Key_Left ? -1 : 1);
    const view = logShown ? lines : list;
    if (!scrolled(view, event.key))
      return false;
    if (view === lines)
      logFollow = lines.atYEnd;
    else
      rowsFollow = list.atYEnd;
    return true;
  }

  function opened(first) {
    logFollow = false;
    lines.waiting = first;
    logShown = true;
    Qt.callLater(lines.follow);
  }

  Keys.onPressed: event => event.accepted = root.pressed(event)
  onQuestionChanged: {
    field.text = "";
    Qt.callLater(settle);
  }
  onRunningChanged: {
    halting = null;
    Qt.callLater(settle);
  }
  onHeldChanged: {
    if (held)
      halting = null;
    else
      Qt.callLater(resume);
  }
  onCodeChanged: Qt.callLater(settle)

  component Mark: Canvas {
    id: mark

    property string kind: "ok"
    property color ink: Theme.text
    property int phase: 0
    readonly property var patterns: ({
        ok: () => true,
        fail: (x, y) => (x + y) % 3 === 0,
        warn: (x, y, cell) => (Math.floor(x / cell) + Math.floor(y / cell)) % 2 === 0,
        busy: (x, y, cell, shift) => (Math.floor((x + shift) / cell) + Math.floor(y / cell)) % 2 === 0
      })

    implicitWidth: Theme.size(12)
    implicitHeight: implicitWidth
    antialiasing: false
    smooth: false
    onKindChanged: requestPaint()
    onInkChanged: requestPaint()
    onPhaseChanged: requestPaint()
    onWidthChanged: requestPaint()
    onPaint: {
      const context = getContext("2d");
      const size = Math.round(width), cell = Math.max(1, Math.round(size / 6));
      const pattern = patterns[kind] || (() => false);
      context.reset();
      context.fillStyle = ink;
      for (let y = 0; y < size; y++)
        for (let x = 0; x < size; x++)
          if (x === 0 || y === 0 || x === size - 1 || y === size - 1 || pattern(x, y, cell, phase))
            context.fillRect(x, y, 1, 1);
    }
  }

  Item {
    id: idle

    focus: true
  }

  TextMetrics {
    id: widest

    font.family: Theme.display
    font.pixelSize: Theme.size(11)
    font.bold: Theme.displayBold
    font.letterSpacing: Theme.size(11) * 0.08
    text: "WARN"
  }

  Timer {
    interval: 120
    running: root.visible && root.task !== null && root.task.busy !== ""
    repeat: true
    onTriggered: root.phase = (root.phase + 1) % 120
  }

  Connections {
    target: root.task

    function onPaged(first) {
      Qt.callLater(root.opened, first);
    }

    function onStoppingChanged() {
      if (root.task.stopping)
        root.halting = null;
    }
  }

  ColumnLayout {
    anchors.fill: parent
    spacing: 16

    Heading {
      Layout.fillWidth: true
      text: root.title
      note: root.condition
    }

    ListView {
      id: list

      function follow() {
        if (root.rowsFollow)
          Qt.callLater(positionViewAtEnd);
      }

      Layout.fillWidth: true
      Layout.fillHeight: true
      clip: true
      boundsBehavior: Flickable.StopAtBounds
      model: root.task ? root.task.rows : null
      onContentHeightChanged: follow()
      onHeightChanged: follow()
      onMovementEnded: root.rowsFollow = atYEnd

      delegate: Item {
        id: row

        required property int index
        required property string kind
        required property string text
        required property string number
        required property string word
        readonly property bool marked: ["ok", "fail", "warn", "skip"].includes(kind)
        readonly property bool failed: kind === "fail"
        readonly property color ink: failed ? Theme.base : Theme.text

        width: ListView.view.width
        height: kind === "rule" ? heading.implicitHeight + (index > 0 ? 16 : 0) : marked ? Theme.size(24) : Math.max(Theme.size(24), line.implicitHeight)

        Heading {
          id: heading

          anchors.left: parent.left
          anchors.right: parent.right
          anchors.bottom: parent.bottom
          visible: row.kind === "rule"
          text: row.text
        }

        Rectangle {
          anchors.fill: parent
          visible: row.marked
          radius: Math.min(Theme.radius, 4)
          color: row.failed ? Theme.text : "transparent"

          RowLayout {
            anchors.fill: parent
            anchors.leftMargin: 8
            anchors.rightMargin: 8
            spacing: 8

            Mark {
              kind: row.kind
              ink: row.ink
            }

            Label {
              Layout.preferredWidth: Math.ceil(widest.advanceWidth)
              text: row.word
              color: row.ink
              font.family: Theme.display
              font.pixelSize: Theme.size(11)
              font.bold: Theme.displayBold
              font.letterSpacing: Theme.size(11) * 0.08
            }

            Label {
              visible: row.number !== ""
              text: row.number
              color: row.failed ? Theme.base : Theme.subtext
            }

            Label {
              Layout.fillWidth: true
              text: row.text
              color: row.ink
              elide: Text.ElideRight
            }
          }
        }

        Label {
          id: line

          anchors.left: parent.left
          anchors.right: parent.right
          anchors.leftMargin: row.kind === "note" ? root.textColumn : 0
          anchors.verticalCenter: parent.verticalCenter
          visible: row.kind === "text" || row.kind === "note"
          text: row.text
          color: row.kind === "note" ? Theme.subtext : Theme.text
          wrapMode: Text.Wrap
        }
      }

      footer: Column {
        width: ListView.view.width

        RowLayout {
          x: 8
          width: parent.width - 16
          height: Theme.size(24)
          visible: root.task !== null && root.task.busy !== ""
          spacing: 8

          Mark {
            kind: "busy"
            phase: root.phase
          }

          Label {
            Layout.preferredWidth: Math.ceil(widest.advanceWidth)
            text: "..."
            font.family: Theme.display
            font.pixelSize: Theme.size(11)
            font.bold: Theme.displayBold
            font.letterSpacing: Theme.size(11) * 0.08
          }

          Label {
            Layout.fillWidth: true
            text: root.task ? root.task.busy : ""
            elide: Text.ElideRight
          }
        }

        Label {
          width: parent.width
          visible: root.reason !== ""
          topPadding: 8
          text: root.reason
          wrapMode: Text.Wrap
        }
      }
    }

    Rectangle {
      Layout.fillWidth: true
      implicitHeight: asked.implicitHeight + 48
      visible: root.question !== null
      radius: Theme.radius
      color: Theme.surface
      border.width: 1
      border.color: Theme.overlay

      ColumnLayout {
        id: asked

        anchors.fill: parent
        anchors.margins: 24
        spacing: 16

        Label {
          Layout.fillWidth: true
          text: root.question ? root.question.text : ""
          wrapMode: Text.Wrap
        }

        Flow {
          id: choices

          function pick(index) {
            const button = buttons.itemAt(index);
            if (button)
              button.forceActiveFocus();
          }

          function step(direction) {
            for (let index = 0; index < buttons.count; index++) {
              if (buttons.itemAt(index).activeFocus) {
                pick(Math.max(0, Math.min(buttons.count - 1, index + direction)));
                return true;
              }
            }
            root.settle();
            return true;
          }

          Layout.fillWidth: true
          visible: root.choosing
          spacing: 8

          Repeater {
            id: buttons

            model: root.choosing ? root.question.items : []

            Button {
              required property string modelData

              label: modelData
              onClicked: root.answered(modelData)
            }
          }
        }

        Field {
          id: field

          Layout.fillWidth: true
          visible: root.question !== null && !root.choosing
          activeFocusOnTab: visible
          secret: root.question !== null && root.question.kind === "password"
          onAccepted: {
            const typed = text;
            text = "";
            root.task.answer(typed);
          }
        }
      }
    }

    Rectangle {
      Layout.fillWidth: true
      Layout.preferredHeight: Math.round(root.height * 0.4)
      visible: root.logShown
      radius: Theme.radius
      color: Theme.surface
      border.width: 1
      border.color: Theme.overlay

      ColumnLayout {
        anchors.fill: parent
        anchors.margins: 16
        spacing: 8

        Heading {
          Layout.fillWidth: true
          text: "Log"
          note: lines.count === 1 ? "1 line" : lines.count + " lines"
        }

        ListView {
          id: lines

          property int waiting: -1

          function follow() {
            if (height <= 0)
              return;
            if (waiting >= 0)
              positionViewAtIndex(waiting, ListView.Beginning);
            else if (root.logFollow)
              positionViewAtEnd();
            waiting = -1;
          }

          Layout.fillWidth: true
          Layout.fillHeight: true
          clip: true
          boundsBehavior: Flickable.StopAtBounds
          model: root.task ? root.task.log : null
          onContentHeightChanged: Qt.callLater(follow)
          onHeightChanged: Qt.callLater(follow)
          onMovementEnded: root.logFollow = atYEnd

          delegate: Label {
            required property var model

            width: ListView.view.width
            text: model.text
            color: Theme.subtext
            elide: Text.ElideRight
            font.pixelSize: Theme.size(12)
          }
        }
      }
    }

    RowLayout {
      Layout.fillWidth: true
      spacing: 16

      Hints {
        Layout.fillWidth: true
        pairs: root.hints
      }

      Button {
        id: ending

        label: root.running ? (root.task.stopping ? "Stopping" : "Stop") : "Close"
        onClicked: {
          if (root.running)
            root.askStop(root.stopQuestion, ["Yes", "No"], "Yes");
          else
            root.closed();
        }
      }
    }
  }
}
