import QtQuick
import Quickshell.Io

QtObject {
  id: root

  property var command: []
  readonly property bool running: child.running
  readonly property ListModel rows: ListModel {}
  property string busy: ""
  readonly property ListModel log: ListModel {}
  property var question: null
  property int code: -1
  property string reason: ""
  property bool stopping: false
  property string partial: ""
  readonly property int logLines: 2000
  readonly property var kinds: ["text", "rule", "ok", "fail", "skip", "warn", "busy", "done", "note", "result", "page", "log", "choose", "input", "password"]
  readonly property var asking: ["choose", "input", "password"]
  readonly property var words: ({
      ok: "PASS",
      fail: "FAIL",
      warn: "WARN",
      skip: "--"
    })
  readonly property var results: ({
      PASS: "ok",
      FAIL: "fail",
      SKIP: "skip"
    })
  readonly property var escapes: ({
      a: "\x07",
      b: "\b",
      f: "\f",
      n: "\n",
      r: "\r",
      t: "\t",
      v: "\v",
      "\\": "\\",
      "\"": "\""
    })
  readonly property var eventPattern: /^([a-z]+)((?: \x22(?:[^\x22\\]+|\\(?:x[0-9a-fA-F]{2}|u[0-9a-fA-F]{4}|U[0-9a-fA-F]{8}|[0-7]{3}|[abfnrtv\\\x22]))*\x22)+)$/
  readonly property var fieldPattern: /\x22(?:[^\x22\\]+|\\.)*\x22/g
  readonly property var escapePattern: /\\(?:x([0-9a-fA-F]{2})|u([0-9a-fA-F]{4})|U([0-9a-fA-F]{8})|([0-7]{3})|(.))|[^\\]+/g
  readonly property var sequencePattern: /\x1b\[[0-9;?]*[ -\/]*[@-~]|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)|\x1b[()*+].|\x1b./g
  readonly property Process child: Process {
    stdinEnabled: true
    stdout: SplitParser {
      onRead: line => root.take(line)
    }
    stderr: SplitParser {
      onRead: line => root.reason = ((root.reason ? root.reason + "\n" : "") + root.clean(line)).slice(-4096)
    }
    onExited: code => root.code = code
    onRunningChanged: {
      if (!running)
        Qt.callLater(root.ended);
    }
  }

  signal paged(int first)

  function start() {
    if (child.running || command.length === 0)
      return;
    rows.clear();
    log.clear();
    partial = "";
    busy = "";
    question = null;
    code = -1;
    reason = "";
    stopping = false;
    child.command = command;
    child.running = true;
  }

  function answer(text) {
    if (!question || !child.running)
      return;
    question = null;
    child.write("answer " + JSON.stringify(String(text)) + "\n");
  }

  function stop() {
    if (!child.running || stopping)
      return;
    stopping = true;
    question = null;
    child.write("stop\n");
  }

  function ended() {
    if (child.running)
      return;
    if (partial !== "")
      logged("\n");
    if (code === -1)
      code = 127;
    busy = "";
    question = null;
    stopping = false;
  }

  function encoded(text) {
    return unescape(encodeURIComponent(text));
  }

  function decoded(bytes) {
    try {
      return decodeURIComponent(escape(bytes));
    } catch (error) {
      return bytes.replace(/[\x80-\xff]/g, "�");
    }
  }

  function clean(line) {
    const kept = line.replace(/\r+$/, "");
    return kept.slice(kept.lastIndexOf("\r") + 1).replace(sequencePattern, "").replace(/\t/g, " ").replace(/[\x00-\x1f\x7f]/g, "");
  }

  function unquoted(quoted) {
    return quoted.slice(1, -1).replace(escapePattern, (all, hex, four, eight, octal, single) => {
      if (hex || octal)
        return String.fromCharCode(parseInt(hex || octal, hex ? 16 : 8));
      if (single)
        return escapes[single];
      return encoded(four || eight ? String.fromCodePoint(parseInt(four || eight, 16)) : all);
    });
  }

  function parsed(line) {
    const found = eventPattern.exec(line);
    if (!found || !kinds.includes(found[1]))
      return null;
    try {
      return {
        kind: found[1],
        fields: found[2].match(fieldPattern).map(unquoted)
      };
    } catch (error) {
      return null;
    }
  }

  function take(line) {
    const found = parsed(line);
    if (!found)
      logged(encoded(line + "\n"));
    else if (found.kind === "log")
      logged(found.fields[0]);
    else
      heard(found.kind, found.fields.map(decoded));
  }

  function heard(kind, texts) {
    if (kind === "busy" || kind === "done")
      busy = kind === "busy" ? clean(texts[0]) : "";
    else if (asking.includes(kind))
      question = {
        kind: kind,
        text: clean(texts[0]),
        items: texts.slice(1).map(clean)
      };
    else if (kind === "page")
      read(texts[0]);
    else if (kind === "result")
      resulted(texts[0], texts[1] || "", texts[2] || "");
    else if (kind === "text")
      texts[0].replace(/\n$/, "").split("\n").forEach(line => add(kind, line, "", ""));
    else
      add(kind, texts[0], "", words[kind] || "");
  }

  function resulted(title, number, word) {
    const kind = results[word] || "warn";
    add(kind, title, number, results[word] ? words[kind] : word);
  }

  function add(kind, text, number, word) {
    rows.append({
      kind: kind,
      text: clean(text),
      number: clean(number),
      word: clean(word)
    });
  }

  function logged(bytes) {
    const lines = (partial + bytes).split("\n");
    partial = lines.pop().slice(-65536);
    for (const line of lines)
      log.append({
        text: clean(decoded(line))
      });
    if (log.count > logLines)
      log.remove(0, log.count - logLines);
  }

  function read(text) {
    const lines = text.replace(/\n$/, "").split("\n");
    logged((partial !== "" ? "\n" : "") + encoded("\n" + lines.join("\n") + "\n"));
    paged(Math.max(0, log.count - lines.length - 1));
  }
}
