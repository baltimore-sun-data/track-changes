import { html } from "es6-string-html-template";
import moment from "moment";

import { getStorageObj } from "./utils.js";

function listRecentSheets() {
  const recentSheetsList = document.querySelector(".recent-sheets-list");
  const lastRefreshObj = getStorageObj("last-refresh") || {};

  // Sort sheets by reverse chron.
  let keys = Object.keys(lastRefreshObj);
  keys.sort((a, b) => lastRefreshObj[a] < lastRefreshObj[b]);

  if (keys.length < 1) {
    return;
  }
  recentSheetsList.innerHTML = "";

  keys.forEach(sheet => {
    let lastRefresh = lastRefreshObj[sheet];

    recentSheetsList.insertAdjacentHTML(
      "beforeend",
      html`
          <li data-date=${lastRefresh}>
            <a href="?sheet=${sheet}">
              Sheet last accessed ${moment(lastRefresh).fromNow()}
            </a>
          </li>
`
    );
  });
}

window.addEventListener("load", listRecentSheets);
