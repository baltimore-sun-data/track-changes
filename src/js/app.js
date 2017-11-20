import { html } from "es6-string-html-template";
import moment from "moment";
import tinysort from "tinysort";

// Convenience extension to NodeList:
NodeList.prototype.addEventListener = function(event, func) {
  this.forEach(function(content, item) {
    content.addEventListener(event, func);
  });
};

const API_URL = "/api";
window.next_poll = 5 * 60 * 1000; // 5 minute default

const table = document.getElementById("xtable");
const tableHead = table.querySelector("thead");
const sortableTh = tableHead.querySelectorAll("th.sortable");
const tableBody = table.querySelector("tbody");

const error = document.getElementById("error");

const refresh = document.querySelectorAll(".refresh-time");
const refreshBtn = document.querySelectorAll(".refresh-btn");

// Global to preserve sorting between refreshes
let sortOptions = {};

// Set up data-indexes for later convenience
tableHead.querySelectorAll("th").forEach((el, i) => {
  el.setAttribute("data-index", i + 1);
});

sortableTh.addEventListener("click", e => {
  const tableHeader = e.target;
  const tableHeaderIndex = tableHeader.getAttribute("data-index");
  const isAscending = tableHeader.getAttribute("data-order") === "asc";
  const order = isAscending ? "desc" : "asc";
  tableHeader.setAttribute("data-order", order);

  sortOptions = {
    selector: `td:nth-child(${tableHeaderIndex})`,
    data: "sort",
    order: order
  };

  tinysort(tableBody.querySelectorAll("tr"), sortOptions);
});

async function updateData() {
  try {
    const rsp = await fetch(API_URL);

    refresh.forEach(el => {
      el.textContent = moment().format("LTS");
    });

    if (!rsp.ok) {
      throw new Error("Could not contact API");
    }

    const body = await rsp.json();

    if (!body.data) {
      throw new Error("No data returned");
    }

    // TODO: Preserve sort order
    tableBody.querySelectorAll("tr").forEach(el => el.remove());

    body.data.forEach(item =>
      tableBody.insertAdjacentHTML(
        "beforeend",
        html`
  <tr data-row-id="${item.id}">
    <td data-sort="${item.display_name}">
      <a href="${item.homepage_url}">${item.display_name}</a>
    </td>
    <td><a href="${item.url}">${item.content}</a></td>
    <td title="${item.twitter_screenname}">
      <a href="https://twitter.com/${item.twitter_screenname}">${
          item.last_tweet
        }</a>
    </td>
    <td
      data-sort="${item.last_change}"
      title="${moment(item.last_change).format("llll")}">
      ${item.last_change ? moment(item.last_change).fromNow() : ""}
    </td>
    <td
      data-sort="${item.last_accessed}"
      title="${moment(item.last_accessed).format("llll")}">
      ${item.last_accessed ? moment(item.last_accessed).fromNow() : ""}
    </td>
    <td data-sort="${item.last_error}">
      ${
        item.error
          ? item.error + " at " + moment(item.last_error).fromNow()
          : ""
      }
    </td>
  </tr>
`
      )
    );
    tinysort(tableBody.querySelectorAll("tr"), sortOptions);

    // Update at the average time between changes for items
    window.next_poll = body.meta.poll_interval / body.data.length;
  } catch (e) {
    error.textContent = `Error returning data: ${e.message}`;
  } finally {
    window.setTimeout(updateData, window.next_poll);
  }
}

window.addEventListener("load", updateData);
refreshBtn.addEventListener("click", updateData);
