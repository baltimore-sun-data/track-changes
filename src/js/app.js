import { html } from "es6-string-html-template";
import moment from "moment";
import tinysort from "tinysort";

const table = document.getElementById("xtable");
const tableHead = table.querySelector("thead");
const sortableTh = tableHead.querySelectorAll("th.sortable");
const tableBody = table.querySelector("tbody");

// Set up data-indexes for later convenience
tableHead.querySelectorAll("th").forEach((el, i) => {
  el.setAttribute("data-index", i + 1);
});

sortableTh.forEach(th =>
  th.addEventListener("click", e => {
    console.log(e);
    const tableHeader = e.target;
    const tableHeaderIndex = tableHeader.getAttribute("data-index");
    const isAscending = tableHeader.getAttribute("data-order") === "asc";
    const order = isAscending ? "desc" : "asc";
    tableHeader.setAttribute("data-order", order);

    tinysort(tableBody.querySelectorAll("tr"), {
      selector: `td:nth-child(${tableHeaderIndex})`,
      data: "time",
      order: order
    });
  })
);

async function updateData() {
  let rsp = await fetch("/api");
  let body = await rsp.json();

  // TODO: Preserve sort order
  tableBody.querySelectorAll("tr").forEach(el => el.remove());

  body.data.forEach(item =>
    tableBody.insertAdjacentHTML(
      "beforeend",
      html`
  <tr data-row-id="${item.id}">
    <td><a href="${item.homepage_url}">${item.display_name}</a></td>
    <td><a href="${item.url}">${item.content}</a></td>
    <td><a href="https://twitter.com/${item.twitter_screenname}">${
        item.last_tweet
      }</a></td>
    <td
      data-time="${item.last_change}"
      title="${moment(item.last_change).format("llll")}">
      ${moment(item.last_change).fromNow()}
    </td>
    <td
      data-time="${item.last_accessed}"
      title="${moment(item.last_accessed).format("llll")}">
      ${moment(item.last_accessed).fromNow()}
    </td>
  </tr>
`
    )
  );

  // Update at the average time between changes for items
  window.next_poll = body.meta.poll_interval / body.data.length;
  window.setTimeout(updateData, window.next_poll);
}

window.addEventListener("load", updateData);
