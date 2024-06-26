let headers = [
    "Symbol",
    "Strategy",
    "TF",
    "Profit/Loss",
    "Win Rate",
    "Trades",
    "avg. t hold",
    "Longs",
    "Shorts",
];

let backtesting_results;

let tableBacktesting = document.querySelector("#resultsTable");

fetch("data/backtesting_results.json")
    .then(function (response) {
        return response.json();
    })
    .then(function (data) {
        data.backtestingResults;
        createTable(data.backtestingResults);
    });

function createTable(backtestingResults) {
    let table = document.createElement("table");
    let headerRow = document.createElement("tr");

    headers.forEach((headerText) => {
        let header = document.createElement("th");
        let textNode = document.createTextNode(headerText);
        header.appendChild(textNode);
        headerRow.appendChild(header);
    });

    table.appendChild(headerRow);

    backtestingResults.forEach((combo) => {
        let row = document.createElement("tr");

        Object.values(combo).forEach((text) => {
            let cell = document.createElement("td");
            let textNode = document.createTextNode(text);
            cell.appendChild(textNode);
            row.appendChild(cell);
        });

        table.appendChild(row);
    });

    tableBacktesting.append(table);
}
