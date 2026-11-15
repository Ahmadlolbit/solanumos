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
  readonly property var specialPattern: /[\x22\\]/g
  readonly property var escapeShape: /^\\(?:[abfnrtv\\\x22]|x[0-9a-fA-F]{2}|u[0-9a-fA-F]{4}|U[0-9a-fA-F]{8}|[0-3][0-7]{2})/
  readonly property var escapePattern: /\\(?:x([0-9a-fA-F]{2})|u([0-9a-fA-F]{4})|U([0-9a-fA-F]{8})|([0-7]{3})|(.))/g
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
    child.write("answer " + JSON.stringify(mended(String(text))) + "\n");
  }

  function mended(text) {
    return text.replace(/[\ud800-\udbff][\udc00-\udfff]|[\ud800-\udfff]/g, pair => pair.length === 2 ? pair : "\ufffd");
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
    return encoded(quoted.slice(1, -1)).replace(escapePattern, (all, hex, four, eight, octal, single) => {
      if (hex || octal)
        return String.fromCharCode(parseInt(hex || octal, hex ? 16 : 8));
      if (single)
        return escapes[single];
      return encoded(String.fromCodePoint(parseInt(four || eight, 16)));
    });
  }

  function closing(text, from) {
    specialPattern.lastIndex = from;
    for (let found = specialPattern.exec(text); found; found = specialPattern.exec(text)) {
      if (found[0] === "\"")
        return found.index;
      const shape = escapeShape.exec(text.slice(found.index, found.index + 10));
      if (!shape)
        return -1;
      specialPattern.lastIndex = found.index + shape[0].length;
    }
    return -1;
  }

  function fields(text) {
    const found = [];
    let at = 0;
    while (at < text.length) {
      const end = text[at] === "\"" ? closing(text, at + 1) : -1;
      if (end < 0)
        return null;
      found.push(text.slice(at, end + 1));
      at = text[end + 1] === " " ? end + 2 : end + 1;
    }
    return found.length > 0 ? found : null;
  }

  function parsed(line) {
    const space = line.indexOf(" ");
    const kind = line.slice(0, space);
    const found = space > 0 && kinds.includes(kind) ? fields(line.slice(space + 1)) : null;
    try {
      return found ? {
        kind: kind,
        fields: found.map(unquoted)
      } : null;
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
