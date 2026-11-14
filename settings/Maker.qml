import QtQuick
import QtQuick.Layouts
import qs.common

Tile {
  id: maker

  property bool open: false
  property bool editing: false
  property string mode: "make"
  property var changes: ({})
  readonly property bool own: Store.themes.some(item => item.name === Store.current && item.image.startsWith(Store.yours.replace(/backgrounds\/$/, "themes/")))
  readonly property var swatches: Store.palette ? Store.palette.colors || [] : []
  readonly property var named: {
    const found = {};
    for (const each of swatches)
      found[each.key] = each.value;
    return found;
  }
  readonly property real swatchWidth: Math.max(Theme.size(232), Math.ceil(Math.max(0, ...swatches.map(each => metrics.advanceWidth(swatchText(each))))) + Theme.size(22) + 24)

  signal pointed

  function swatchText(each) {
    return each.key.replace(/_/g, " ") + " " + each.value + "  " + each.contrast.toFixed(1) + ":1";
  }

  function hex(text) {
    return /^#[0-9A-Fa-f]{6}$/.test(text.trim());
  }

  function refresh() {
    if (!open)
      return;
    if (mode === "make" && hex(background.text) && hex(foreground.text) && hex(accent.text))
      Store.preview(["solanum-theme-set", "--palette", background.text.trim(), foreground.text.trim(), accent.text.trim()]);
    else if (mode === "change" && Store.current)
      Store.preview(["solanum-theme-set", "--palette", Store.current].concat(Object.keys(changes).map(key => key + "=" + changes[key])));
  }

  function save() {
    const name = nameField.text.trim();
    if (mode === "make" && name && hex(background.text) && hex(foreground.text) && hex(accent.text))
      Store.act("make", ["solanum-theme-set", "--make", name, background.text.trim(), foreground.text.trim(), accent.text.trim()], "themes");
    else if (mode === "change" && own && !name)
      Store.act("change", ["solanum-theme-set", "--change", Store.current].concat(Object.keys(changes).map(key => key + "=" + changes[key])), "themes");
    else if (mode === "change" && name)
      Store.act("copy", ["solanum-theme-set", "--copy", Store.current, name].concat(Object.keys(changes).map(key => key + "=" + changes[key])), "themes");
    editing = false;
  }

  function activate() {
    if (!open) {
      open = true;
      refresh();
      return true;
    }
    editing = true;
    nameField.focusInput();
    return true;
  }

  function step(direction) {
    if (!open)
      return false;
    mode = direction < 0 ? "make" : "change";
    Store.palette = null;
    refresh();
    return true;
  }

  title: "Your own theme"
  hint: !open ? (current ? "Enter opens it" : "From three colours, or any theme's colours changed") : current ? "Left: three colours · Right: change this theme's colours · Enter to type · Esc stops" : ""

  FontMetrics {
    id: metrics

    font.family: Theme.font
    font.pixelSize: Theme.size(12)
  }

  ColumnLayout {
    width: parent.width
    spacing: 14

    Label {
      Layout.fillWidth: true
      visible: !maker.open
      text: "Make a theme from a background, a foreground and an accent; the rest is worked out and shown before saving. Or change any colour of the theme in use as your own copy."
      color: Theme.subtext
      wrapMode: Text.Wrap
    }

    RowLayout {
      visible: maker.open
      spacing: 8

      Choice {
        label: "From three colours"
        selected: maker.mode === "make"
        onClicked: maker.step(-1)
      }

      Choice {
        label: "Change " + (Store.current ? Store.current.replace(/-/g, " ") : "this theme") + "'s colours"
        selected: maker.mode === "change"
        onClicked: maker.step(1)
      }
    }

    GridLayout {
      visible: maker.open
      columns: 2
      columnSpacing: 14
      rowSpacing: 8

      Label {
        text: maker.mode === "change" && maker.own ? "Name, or empty to change it" : "Name"
        color: Theme.subtext
      }

      Field {
        id: nameField

        placeholder: maker.mode === "make" ? "my-theme" : "my-" + Store.current
        onAccepted: maker.save()
      }

      Label {
        visible: maker.mode === "make"
        text: "Background"
        color: Theme.subtext
      }

      Field {
        id: background

        visible: maker.mode === "make"
        text: "#002941"
        onTextChanged: maker.refresh()
        onAccepted: maker.save()
      }

      Label {
        visible: maker.mode === "make"
        text: "Foreground"
        color: Theme.subtext
      }

      Field {
        id: foreground

        visible: maker.mode === "make"
        text: "#b6dfed"
        onTextChanged: maker.refresh()
        onAccepted: maker.save()
      }

      Label {
        visible: maker.mode === "make"
        text: "Accent"
        color: Theme.subtext
      }

      Field {
        id: accent

        visible: maker.mode === "make"
        text: "#fd6e6e"
        onTextChanged: maker.refresh()
        onAccepted: maker.save()
      }
    }

    Rectangle {
      visible: maker.open && !!Store.palette
      Layout.fillWidth: true
      Layout.preferredHeight: sample.implicitHeight + 28
      color: maker.named.background || Theme.base
      radius: Theme.radius
      border.width: 1
      border.color: Theme.overlay

      Column {
        id: sample

        anchors.left: parent.left
        anchors.right: parent.right
        anchors.top: parent.top
        anchors.margins: 14
        spacing: 6

        Text {
          text: "Text in the foreground reads like this."
          color: maker.named.foreground || Theme.text
          font.family: Theme.font
          font.pixelSize: Theme.fontSize
        }

        Text {
          text: "The accent marks what's chosen."
          color: maker.named.accent || Theme.accent
          font.family: Theme.font
          font.pixelSize: Theme.fontSize
        }

        Rectangle {
          width: chosen.implicitWidth + 16
          height: chosen.implicitHeight + 6
          color: maker.named.selection || Theme.overlay

          Text {
            id: chosen

            anchors.centerIn: parent
            text: "Selected text stays readable."
            color: maker.named.foreground || Theme.text
            font.family: Theme.font
            font.pixelSize: Theme.fontSize
          }
        }

        Row {
          spacing: 4

          Repeater {
            model: ["red", "yellow", "orange", "green", "cyan", "blue", "magenta", "bright_red", "bright_yellow", "bright_green", "bright_cyan", "bright_blue", "bright_magenta"]

            Rectangle {
              required property string modelData

              width: Theme.size(26)
              height: Theme.size(18)
              color: maker.named[modelData] || "transparent"
            }
          }
        }
      }
    }

    Flow {
      id: colours

      visible: maker.open && !!Store.palette
      Layout.fillWidth: true
      Layout.maximumWidth: parent.width
      spacing: 8

      Repeater {
        model: maker.swatches

        Rectangle {
          id: swatch

          required property var modelData

          width: Math.min(colours.width, maker.swatchWidth)
          height: Math.max(Theme.size(34), label.implicitHeight + 12)
          radius: Theme.radius
          color: Theme.surface

          Rectangle {
            id: dot

            anchors.left: parent.left
            anchors.verticalCenter: parent.verticalCenter
            anchors.leftMargin: 6
            width: Theme.size(22)
            height: width
            color: swatch.modelData.value
            border.width: 1
            border.color: Theme.overlay
          }

          Label {
            id: label

            anchors.left: dot.right
            anchors.right: parent.right
            anchors.verticalCenter: parent.verticalCenter
            anchors.leftMargin: 8
            anchors.rightMargin: 6
            text: maker.swatchText(swatch.modelData)
            color: swatch.modelData.contrast < (swatch.modelData.key === "accent" ? 3 : 4.5) && swatch.modelData.against === "background" && !swatch.modelData.key.includes("background") ? Theme.critical : Theme.text
            font.pixelSize: Theme.size(12)
            wrapMode: Text.Wrap
          }

          MouseArea {
            anchors.fill: parent
            enabled: maker.mode === "change"
            cursorShape: Qt.PointingHandCursor
            onClicked: {
              maker.pointed();
              picked.key = swatch.modelData.key;
              picked.text = swatch.modelData.value;
              maker.editing = true;
              picked.focusInput();
            }
          }
        }
      }
    }

    RowLayout {
      visible: maker.open && maker.mode === "change"
      spacing: 10

      Label {
        text: picked.key ? "New " + picked.key.replace(/_/g, " ") : "Click a colour to change it"
        color: Theme.subtext
      }

      Field {
        id: picked

        property string key: ""

        visible: !!key
        implicitWidth: Theme.size(120)
        onAccepted: {
          if (maker.hex(text)) {
            const next = Object.assign({}, maker.changes);
            next[key] = text.trim().toLowerCase();
            maker.changes = next;
            maker.refresh();
          }
        }
      }
    }

    Repeater {
      model: maker.open && Store.palette ? Store.palette.fixes || [] : []

      Label {
        required property string modelData

        Layout.fillWidth: true
        text: "\uf071  " + modelData
        color: Theme.warning
        wrapMode: Text.Wrap
      }
    }

    Choice {
      visible: maker.open
      label: maker.mode === "make" ? "Save it as your own theme" : maker.own && !nameField.text.trim() ? "Save the changes" : "Save as your own copy"
      icon: "\uf0c7"
      selected: true
      onClicked: maker.save()
    }
  }
}
