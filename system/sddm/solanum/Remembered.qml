import QtCore

Settings {
  property bool larger: false
  property bool contrast: false
  property bool keyboard: false

  location: StandardPaths.writableLocation(StandardPaths.GenericConfigLocation) + "/solanum-login.conf"
  category: "accessibility"
}
