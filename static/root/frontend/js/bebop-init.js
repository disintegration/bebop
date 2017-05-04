marked.setOptions({
  sanitize: true,
  breaks: true,
});

Vue.filter("formatTime", function(value) {
  if (value) {
    return moment(String(value)).format("MMMM Do YYYY, hh:mm");
  }
});

Vue.filter("formatTimeAgo", function(value) {
  if (value) {
    return moment(String(value)).fromNow();
  }
});

Vue.filter("capitalize", function(value) {
  if (value) {
    value = String(value);
    return value[0].toUpperCase() + value.slice(1);
  }
});

function getPagination(curPage, lastPage) {
  var pagination = [];
  var lr = 2;

  pagination.push(1);

  if (curPage - lr > 2) {
    pagination.push("...");
  }

  for (var p = curPage - lr; p <= curPage + lr; p++) {
    if (p > 1 && p < lastPage) {
      pagination.push(p);
    }
  }

  if (curPage + lr < lastPage - 1) {
    pagination.push("...");
  }

  if (lastPage > 1) {
    pagination.push(lastPage);
  }

  return pagination;
}
