pragma Singleton

import Quickshell

Singleton {
  property bool used: true

  function joins(search, text) {
    return !!text && !/[\u0000-\u001f\u007f-\u009f]/.test(text) && !!(search || text.trim());
  }
}
