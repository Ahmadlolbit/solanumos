import QtQuick
import QtQuick.Layouts
import Quickshell
import Quickshell.Services.Pipewire
import qs.common

ColumnLayout {
  id: root

  readonly property PwNode sink: Pipewire.defaultAudioSink
  readonly property PwNode source: Pipewire.defaultAudioSource
  readonly property bool muted: sink && sink.audio ? sink.audio.muted : false
  readonly property real volume: sink && sink.audio ? sink.audio.volume : 0
  readonly property bool micMuted: source && source.audio ? source.audio.muted : false
  readonly property real micVolume: source && source.audio ? source.audio.volume : 0

  signal done

  function icon() {
    if (muted || volume === 0)
      return "\u{f0581}";
    if (volume < 0.34)
      return "\u{f057f}";
    if (volume < 0.67)
      return "\u{f0580}";
    return "\u{f057e}";
  }

  function devices(sinks) {
    return Pipewire.nodes.values.filter(node => node.audio && !node.isStream && node.isSink === sinks);
  }

  function title(node) {
    return node.description || node.nickname || node.name;
  }

  function streams() {
    return Pipewire.nodes.values.filter(node => node.audio && node.isStream && node.isSink);
  }

  function appName(node) {
    return node.properties["application.name"] || node.description || node.name;
  }

  function mute(node) {
    if (node && node.audio)
      node.audio.muted = !node.audio.muted;
  }

  function setVolume(node, level) {
    if (node && node.audio)
      node.audio.volume = level;
  }

  spacing: 10

  PwObjectTracker {
    objects: [root.sink, root.source]
  }

  PwObjectTracker {
    objects: root.streams()
  }

  Label {
    Layout.preferredWidth: Theme.size(300)
    text: "Output"
    font.bold: true
  }

  RowLayout {
    Layout.fillWidth: true
    spacing: 10

    Button {
      Layout.preferredWidth: Theme.size(30)
      label: root.icon()
      onClicked: root.mute(root.sink)
    }

    Slider {
      Layout.fillWidth: true
      value: root.volume
      onMoved: level => root.setVolume(root.sink, level)
    }

    Label {
      Layout.preferredWidth: Theme.size(36)
      horizontalAlignment: Text.AlignRight
      text: Math.round(root.volume * 100) + "%"
    }
  }

  Repeater {
    model: root.devices(true)

    Choice {
      required property PwNode modelData

      Layout.fillWidth: true
      icon: "\u{f04c3}"
      label: root.title(modelData)
      selected: modelData === root.sink
      onClicked: Pipewire.preferredDefaultAudioSink = modelData
    }
  }

  Label {
    visible: root.streams().length > 0
    text: "Apps"
    font.bold: true
  }

  Repeater {
    model: root.streams()

    RowLayout {
      id: app

      required property PwNode modelData
      readonly property bool muted: modelData.audio ? modelData.audio.muted : false
      readonly property real volume: modelData.audio ? modelData.audio.volume : 0

      Layout.fillWidth: true
      spacing: 10

      Button {
        Layout.preferredWidth: Theme.size(30)
        label: app.muted ? "\u{f075b}" : "\u{f075a}"
        onClicked: root.mute(app.modelData)
      }

      Label {
        Layout.preferredWidth: Theme.size(90)
        text: root.appName(app.modelData)
        elide: Text.ElideRight
      }

      Slider {
        Layout.fillWidth: true
        value: app.volume
        onMoved: level => root.setVolume(app.modelData, level)
      }

      Label {
        Layout.preferredWidth: Theme.size(36)
        horizontalAlignment: Text.AlignRight
        text: Math.round(app.volume * 100) + "%"
      }
    }
  }

  Label {
    text: "Input"
    font.bold: true
  }

  RowLayout {
    Layout.fillWidth: true
    spacing: 10

    Button {
      Layout.preferredWidth: Theme.size(30)
      label: root.micMuted ? "\u{f036d}" : "\u{f036c}"
      onClicked: root.mute(root.source)
    }

    Slider {
      Layout.fillWidth: true
      value: root.micVolume
      onMoved: level => root.setVolume(root.source, level)
    }

    Label {
      Layout.preferredWidth: Theme.size(36)
      horizontalAlignment: Text.AlignRight
      text: Math.round(root.micVolume * 100) + "%"
    }
  }

  Repeater {
    model: root.devices(false)

    Choice {
      required property PwNode modelData

      Layout.fillWidth: true
      icon: "\u{f036c}"
      label: root.title(modelData)
      selected: modelData === root.source
      onClicked: Pipewire.preferredDefaultAudioSource = modelData
    }
  }

  Choice {
    Layout.fillWidth: true
    icon: "\u{f013}"
    label: "More settings"
    onClicked: {
      root.done();
      Quickshell.execDetached(["pavucontrol"]);
    }
  }
}
