import { html } from "es6-string-html-template";
import moment from "moment";
import tinysort from "tinysort";

import { getStorageObj, setStorageObj } from "./utils.js";

// Constant selectors
const tables = document.querySelectorAll(".item-table");
const sortableTh = document.querySelectorAll("th.sortable");
const activeTable = document.querySelector("#active-table");
const processedTable = document.querySelector("#processed-table");

const sheetTitle = document.querySelector(".header-sheet-title");

const error = document.getElementById("error");

const refresh = document.querySelectorAll(".refresh-time");
const refreshBtn = document.querySelectorAll(".refresh-btn");

const sheetBtn = document.querySelectorAll(".sheet-btn");

// Global values
const apiUrl = `/api/sheet/${window.trackChanges.sheetID}`;
const apiOptions = !window.trackChanges.basicAuthHeader
  ? {}
  : {
      method: "GET",
      headers: {
        Authorization: `Basic ${window.trackChanges.basicAuthHeader}`
      }
    };

const processedIDsKey = `processed-ids-${window.trackChanges.sheetID}`;
const processedIDs = getStorageObj(processedIDsKey) || {};
const apiData = {
  "active-table": [],
  "processed-table": []
};

window.next_poll = 5 * 60 * 1000; // 5 minute default

// Set up data-indexes for later convenience
tables.forEach(table => {
  // Keep sort options on node to preserve between refreshes
  table.sortOptions = {};
  table.querySelectorAll("thead th").forEach((el, i) => {
    el.setAttribute("data-index", i + 1);
  });

  table.querySelector(".table-group-btn").addEventListener("click", e => {
    e.target.checked = false;
    const isProcessed = table.id === "processed-table";
    const dstTableID = isProcessed ? "active-table" : "processed-table";
    Array.prototype.push.apply(apiData[dstTableID], apiData[table.id]);
    apiData[table.id] = [];

    apiData[dstTableID].forEach(row => {
      processedIDs[row.id] = !isProcessed;
    });

    setStorageObj(processedIDsKey, processedIDs);

    displayData(activeTable, apiData["active-table"]);
    displayData(processedTable, apiData["processed-table"]);
  });
});

function sortColumn(e) {
  const tableHeader = e.target;
  const table = tableHeader.closest("table");
  const tableHeaderIndex = tableHeader.getAttribute("data-index");
  const isDescending = tableHeader.getAttribute("data-order") === "desc";
  const order = isDescending ? "asc" : "desc";
  tableHeader.setAttribute("data-order", order);

  // Change highlight
  const selector = `td:nth-child(${tableHeaderIndex})`;

  table.sortOptions = {
    selector: selector,
    data: "sort",
    order: order
  };

  table
    .querySelectorAll(".sort-col")
    .forEach(el => el.classList.remove("sort-col"));
  tableHeader.classList.add("sort-col");
  table.querySelectorAll(selector).forEach(el => el.classList.add("sort-col"));
  tinysort(table.querySelectorAll("tbody tr"), table.sortOptions);
}

async function updateData() {
  try {
    let rsp;

    try {
      rsp = await fetch(apiUrl, apiOptions);
    } catch (e) {
      throw new Error(`Problem connecting to API: ${e.message}`);
    }

    if (!rsp.ok) {
      throw new Error("Unexpected response from API");
    }

    const body = await rsp.json();

    if (!body.data) {
      throw new Error("No data returned");
    }

    sheetTitle.innerText = `(${body.meta.sheet_title})`;

    // Save this sheet for listing on homepage
    const now = moment();
    const recentSheetsObj = getStorageObj("recent-sheets") || {};
    recentSheetsObj[window.trackChanges.sheetID] = {
      time: now,
      title: body.meta.sheet_title
    };
    setStorageObj("recent-sheets", recentSheetsObj);

    refresh.forEach(el => {
      el.textContent = now.format("LTS");
    });

    apiData["active-table"] = [];
    apiData["processed-table"] = [];

    body.data.forEach(row => {
      if (processedIDs[row.id]) {
        apiData["processed-table"].push(row);
      } else {
        apiData["active-table"].push(row);
      }
    });

    displayData(activeTable, apiData["active-table"]);
    displayData(processedTable, apiData["processed-table"]);

    error.classList.add("display-none");

    // Update a bit more frequently than the poller
    window.next_poll = body.meta.poll_interval / 4;
  } catch (e) {
    error.classList.remove("display-none");
    error.textContent = `Error returning data: ${e.message}`;
    throw e;
  } finally {
    window.setTimeout(updateData, window.next_poll);
  }
}

function displayData(table, data) {
  // New table contents
  let tableBody = document.createElement("tbody");

  data.forEach(item =>
    tableBody.insertAdjacentHTML(
      "beforeend",
      html`
  <tr data-row-id="${item.id}">
    <td class="table-check">
      <input type="checkbox" class="table-group-btn">
    </td>
    <td data-sort="${item.display_name}">
      <a href="${item.homepage_url}" target="_blank">${item.display_name}</a>
    </td>
    <td><a href="${item.url}" target="_blank">${item.content}</a></td>
    <td title="${item.twitter_screenname}">
      <a
        href="https://twitter.com/${item.twitter_screenname}"
        target="_blank"
      >${item.last_tweet}</a>
    </td>
    <td
      class="table-time"
      data-sort="${item.last_change}"
      title="${moment(item.last_change).format("llll")}">
      ${item.last_change ? moment(item.last_change).fromNow() : ""}
    </td>
    <td
      class="table-time"
      data-sort="${item.last_accessed}"
      title="${moment(item.last_accessed).format("llll")}">
      ${item.last_accessed ? moment(item.last_accessed).fromNow() : ""}
    </td>
    <td
      class="table-error"
      data-sort="${item.last_error}">
      ${
        item.error
          ? item.error + " at " + moment(item.last_error).fromNow()
          : ""
      }
      ${
        item.twitter_error
          ? item.twitter_error + " at " + moment(item.last_error).fromNow()
          : ""
      }
    </td>
  </tr>
`
    )
  );

  tableBody
    .querySelectorAll(".table-group-btn")
    .forEach(el => el.addEventListener("click", changeTableGroup));

  tableBody
    .querySelectorAll(table.sortOptions.selector)
    .forEach(el => el.classList.add("sort-col"));
  tinysort(tableBody.querySelectorAll("tr"), table.sortOptions);

  // Swap in the new table contents
  let oldBody = table.querySelector("tbody");
  oldBody.parentNode.replaceChild(tableBody, oldBody);
}

function changeTableGroup(e) {
  const srcTable = e.target.closest("table");
  const isProcessed = srcTable.id === "processed-table";
  const dstTableID = isProcessed ? "active-table" : "processed-table";
  const rowID = e.target.closest("tr").attributes["data-row-id"].value;
  const srcIdx = apiData[srcTable.id].findIndex(row => row.id === rowID);
  apiData[dstTableID].push(apiData[srcTable.id][srcIdx]);
  apiData[srcTable.id].splice(srcIdx, 1);

  displayData(activeTable, apiData["active-table"]);
  displayData(processedTable, apiData["processed-table"]);

  // Save in localStorage for page reloads
  processedIDs[rowID] = !isProcessed;
  setStorageObj(processedIDsKey, processedIDs);
}

async function updateSheet() {
  try {
    let rsp;
    let opts = Object.assign({}, apiOptions, { method: "POST" });

    try {
      rsp = await fetch(apiUrl, opts);
    } catch (e) {
      throw new Error(`Problem connecting to API: ${e.message}`);
    }

    if (!rsp.ok) {
      throw new Error("Could not contact API");
    }

    return updateData();
  } catch (e) {
    error.classList.remove("display-none");
    error.textContent = `Error returning data: ${e.message}`;
  }
}

window.addEventListener("load", updateData);
refreshBtn.addEventListener("click", updateData);
sheetBtn.addEventListener("click", updateSheet);
sortableTh.addEventListener("click", sortColumn);
