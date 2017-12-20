import { html } from "es6-string-html-template";
import moment from "moment";

import { appProps } from "./storage.js";

function listRecentSheets() {
  const recentSheetsList = document.querySelector(".recent-sheets-list");
  const recentSheets = appProps.recentSheets;

  if (recentSheets.length < 1) {
    return;
  }

  recentSheetsList.innerHTML = "";

  recentSheets.forEach(obj => {
    recentSheetsList.insertAdjacentHTML(
      "beforeend",
      html`
          <li data-date=${obj.time}>
            <a href="?sheet=${obj.sheetID}">
              Sheet "${obj.title}" last accessed ${moment(obj.time).fromNow()}
            </a>
          </li>
`
    );
  });
}

window.addEventListener("load", listRecentSheets);
