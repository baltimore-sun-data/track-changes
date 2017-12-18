import { html } from "es6-string-html-template";
import moment from "moment";
import tinysort from "tinysort";

import { getStorageObj, setStorageObj } from "./utils.js";

const apiUrl = `/api/sheet/${window.trackChanges.sheetID}`;
const apiOptions = !window.trackChanges.basicAuthHeader
  ? {}
  : {
      method: "GET",
      headers: {
        Authorization: `Basic ${window.trackChanges.basicAuthHeader}`
      }
    };
window.next_poll = 5 * 60 * 1000; // 5 minute default

const table = document.getElementById("xtable");
const tableHead = table.querySelector("thead");
const sortableTh = tableHead.querySelectorAll("th.sortable");

const sheetTitle = document.querySelector(".header-sheet-title");

const error = document.getElementById("error");

const refresh = document.querySelectorAll(".refresh-time");
const refreshBtn = document.querySelectorAll(".refresh-btn");

const sheetBtn = document.querySelectorAll(".sheet-btn");

// Global to preserve sorting between refreshes
let sortOptions = {};

// Set up data-indexes for later convenience
tableHead.querySelectorAll("th").forEach((el, i) => {
  el.setAttribute("data-index", i + 1);
});

sortableTh.addEventListener("click", e => {
  const tableHeader = e.target;
  const tableHeaderIndex = tableHeader.getAttribute("data-index");
  const isDescending = tableHeader.getAttribute("data-order") === "desc";
  const order = isDescending ? "asc" : "desc";
  tableHeader.setAttribute("data-order", order);

  // Change highlight
  let selector = `td:nth-child(${tableHeaderIndex})`;

  sortOptions = {
    selector: selector,
    data: "sort",
    order: order
  };

  table
    .querySelectorAll(".sort-col")
    .forEach(el => el.classList.remove("sort-col"));
  tableHeader.classList.add("sort-col");
  table.querySelectorAll(selector).forEach(el => el.classList.add("sort-col"));
  tinysort(table.querySelectorAll("tbody tr"), sortOptions);
});

async function updateData() {
  try {
    let rsp;

    try {
      rsp = await fetch(apiUrl, apiOptions);
    } catch (e) {
      throw new Error(`Problem connecting to API: ${e.message}`);
    }

    if (!rsp.ok) {
      throw new Error("Could not contact API");
    }

    const body = await rsp.json();

    if (!body.data) {
      throw new Error("No data returned");
    }

    sheetTitle.innerText = `(${body.meta.sheet_title})`;

    // Save this sheet for listing on homepage
    const now = moment();
    let recentSheetsObj = getStorageObj("recent-sheets") || {};
    recentSheetsObj[window.trackChanges.sheetID] = {
      time: now,
      title: body.meta.sheet_title
    };
    setStorageObj("recent-sheets", recentSheetsObj);

    refresh.forEach(el => {
      el.textContent = now.format("LTS");
    });

    // New table contents
    let tableBody = document.createElement("tbody");

    body.data.forEach(item =>
      tableBody.insertAdjacentHTML(
        "beforeend",
        html`
  <tr data-row-id="${item.id}">
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
      .querySelectorAll(sortOptions.selector)
      .forEach(el => el.classList.add("sort-col"));
    tinysort(tableBody.querySelectorAll("tr"), sortOptions);

    // Swap in the new table contents
    let oldBody = table.querySelector("tbody");
    oldBody.parentNode.replaceChild(tableBody, oldBody);

    error.classList.add("display-none");

    // Update a bit more frequently than the poller
    window.next_poll = body.meta.poll_interval / 4;
  } catch (e) {
    error.classList.remove("display-none");
    error.textContent = `Error returning data: ${e.message}`;
  } finally {
    window.setTimeout(updateData, window.next_poll);
  }
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
