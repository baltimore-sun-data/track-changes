// Convenience extension to NodeList:
NodeList.prototype.addEventListener = function(event, func) {
  this.forEach(function(content, item) {
    content.addEventListener(event, func);
  });
};

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

export { getStorageObj, setStorageObj };
