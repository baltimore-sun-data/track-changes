import { html } from "es6-string-html-template";
import moment from "moment";

import { getStorageObj } from "./utils.js";

function listRecentSheets() {
  const recentSheetsList = document.querySelector(".recent-sheets-list");
  const recentSheetsObj = getStorageObj("recent-sheets") || {};

  // Sort sheets by reverse chron.
  let keys = Object.keys(recentSheetsObj);
  keys.sort((a, b) => recentSheetsObj[a].time < recentSheetsObj[b].time);

  if (keys.length < 1) {
    return;
  }
  recentSheetsList.innerHTML = "";

  keys.forEach(sheetID => {
    let sheet = recentSheetsObj[sheetID];

    recentSheetsList.insertAdjacentHTML(
      "beforeend",
      html`
          <li data-date=${sheet.time}>
            <a href="?sheet=${sheetID}">
              Sheet "${sheet.title}" last accessed ${moment(
        sheet.time
      ).fromNow()}
            </a>
          </li>
`
    );
  });
}

window.addEventListener("load", listRecentSheets);
