import QtQuick
import Quickshell.Io

Item {
  id: player

  property string name: ""
  property int cols: 1
  property int rows: 1
  property bool resting: false
  property string frame: ""
  property string asked: ""
  property int shown: -1
  property int generation: 0
  property int soups: 0
  property int population: 0
  readonly property string grid: name + " " + cols + " " + rows
  readonly property bool ready: picture.status === Image.Ready
  readonly property bool playing: shown >= 0 && visible && !resting

  function load() {
    if (name === "" || cols < 2 || rows < 2 || asked === grid)
      return;
    asked = grid;
    shown = -1;
    frame = "";
    server.running = false;
    server.command = ["solanum-background", "--play", name, String(cols), String(rows)];
    server.running = true;
  }

  function take(line) {
    const fields = line.split(" ");
    if (fields.length !== 4) {
      frame = line;
      return;
    }
    generation = Number(fields[1]);
    soups = Number(fields[2]);
    population = Number(fields[3]);
    shown = Number(fields[0]);
  }

  function tell(word) {
    if (server.running)
      server.write(word + "\n");
  }

  width: cols
  height: rows
  transformOrigin: Item.TopLeft
  onGridChanged: settle.restart()
  Component.onCompleted: settle.restart()

  Image {
    id: picture

    anchors.fill: parent
    source: player.frame && player.shown >= 0 ? "file://" + player.frame + "-" + player.shown % 2 + ".ppm?" + player.shown : ""
    cache: false
    smooth: false
    asynchronous: false
  }

  Process {
    id: server

    stdinEnabled: true
    stdout: SplitParser {
      onRead: line => player.take(line)
    }
  }

  Timer {
    id: settle

    interval: 250
    onTriggered: player.load()
  }

  Timer {
    interval: 125
    running: player.playing
    repeat: true
    onTriggered: player.tell("step")
  }

  Connections {
    target: Theme

    function onBaseChanged() {
      player.tell("paint");
    }

    function onTextChanged() {
      player.tell("paint");
    }
  }
}
