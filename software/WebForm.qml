import QtQuick
import qs.common

FocusScope {
  id: form

  readonly property var fields: [name, address, icon]

  signal add(string name, string address, string icon)
  signal cancel

  function start() {
    name.text = "";
    address.text = "https://";
    icon.text = "";
    name.focusInput();
  }

  function submit() {
    if (name.text.trim() !== "" && address.text.trim() !== "")
      form.add(name.text.trim(), address.text.trim(), icon.text.trim());
  }

  function hop(event, at) {
    const step = event.key === Qt.Key_Tab ? 1 : event.key === Qt.Key_Backtab ? -1 : 0;
    if (step === 0)
      return false;
    fields[(at + step + fields.length) % fields.length].focusInput();
    return true;
  }

  implicitHeight: column.implicitHeight

  Column {
    id: column

    width: parent.width
    spacing: 8

    Heading {
      width: parent.width
      text: "Name"
    }

    Field {
      id: name

      width: Math.min(parent.width, Theme.size(480))
      placeholder: "Mail"
      onAccepted: address.focusInput()
      Keys.onPressed: event => event.accepted = form.hop(event, 0)
    }

    Heading {
      width: parent.width
      text: "Address"
    }

    Field {
      id: address

      width: Math.min(parent.width, Theme.size(480))
      placeholder: "https://mail.example.org"
      onAccepted: form.submit()
      Keys.onPressed: event => event.accepted = form.hop(event, 1)
    }

    Heading {
      width: parent.width
      text: "Icon"
      note: "A picture file or an icon's name. Left empty, it takes the site's own."
    }

    Field {
      id: icon

      width: Math.min(parent.width, Theme.size(480))
      placeholder: "Empty takes the site's own"
      onAccepted: form.submit()
      Keys.onPressed: event => event.accepted = form.hop(event, 2)
    }

    Item {
      width: 1
      height: 8
    }

    Row {
      spacing: 16

      Button {
        label: "Add"
        primary: true
        onClicked: form.submit()
      }

      Button {
        label: "Cancel"
        onClicked: form.cancel()
      }
    }
  }
}
