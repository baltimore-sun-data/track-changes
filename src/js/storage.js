import moment from "moment";

function getStorageObj(key) {
  const blob = localStorage.getItem(key);
  if (!blob) {
    return null;
  }

  return JSON.parse(blob);
}

function setStorageObj(key, obj) {
  const blob = JSON.stringify(obj);
  localStorage.setItem(key, blob);
}

const appProps = new class {
  constructor() {
    this.recentSheetsKey = "recent-sheets";
    this.recentSheetsObj = getStorageObj(this.recentSheetsKey) || {};
    // Listens for changes by other tabs.
    window.addEventListener("storage", e => {
      this.recentSheetsObj = getStorageObj(this.recentSheetsKey) || {};
    });
  }

  addRecentSheet(sheetID, pageTitle) {
    const now = moment();
    this.recentSheetsObj[sheetID] = {
      time: now,
      title: pageTitle
    };
    setStorageObj("recent-sheets", this.recentSheetsObj);
  }

  get recentSheets() {
    // Sort sheets by reverse chron.
    let keys = Object.keys(this.recentSheetsObj);
    keys.sort(
      (a, b) => this.recentSheetsObj[a].time < this.recentSheetsObj[b].time
    );
    return keys.map(key => ({
      sheetID: key,
      time: this.recentSheetsObj[key].time,
      title: this.recentSheetsObj[key].title
    }));
  }
}();

export { getStorageObj, setStorageObj, appProps };
