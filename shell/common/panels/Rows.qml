import QtQuick

ListView {
  id: root

  function follow() {
    if (currentIndex < 0 && count > 0)
      currentIndex = 0;
    positionViewAtIndex(currentIndex, ListView.Contain);
    if (activeFocus && currentItem)
      currentItem.forceActiveFocus();
  }

  visible: count > 0
  clip: true
  activeFocusOnTab: true
  onCurrentItemChanged: Qt.callLater(follow)
  onActiveFocusChanged: Qt.callLater(follow)
}
