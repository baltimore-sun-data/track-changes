import tinysort from "tinysort";

const table = document.getElementById("xtable");
const tableHead = table.querySelector("thead");
const tableHeaders = tableHead.querySelectorAll("th");
const tableBody = table.querySelector("tbody");

// Set up data-indexes for later convenience
tableHeaders.forEach((el, i) => {
  el.setAttribute("data-index", i + 1);
});

tableHead.addEventListener("click", e => {
  let tableHeader = e.target;
  while (tableHeader.nodeName !== "TH") {
    tableHeader = tableHeader.parentNode;
  }

  const tableHeaderIndex = tableHeader.getAttribute("data-index");
  const isAscending = tableHeader.getAttribute("data-order") === "asc";
  const order = isAscending ? "desc" : "asc";
  tableHeader.setAttribute("data-order", order);

  tinysort(tableBody.querySelectorAll("tr"), {
    selector: `td:nth-child(${tableHeaderIndex})`,
    order: order
  });
});

const updateData = e => {
  fetch("/api")
    .then(rsp => rsp.json())
    .then(body =>
      body.data.forEach(item => {
        tableBody.insertAdjacentHTML(
          "beforeend",
          `
      <tr data-row-id="${item.id}">
        <td><a href="${item.homepage_url}">${item.display_name}</a></td>
        <td><a href="${item.url}">${item.content}</a></td>
        <td><a href="https://twitter.com/${item.twitter_screenname}">${
            item.last_tweet
          }</a></td>
        <td>${item.last_change}</td>
        <td>${item.last_accessed}</td>
      </tr>
    `
        );
      })
    );
};

window.addEventListener("load", updateData);
