export async function readCsv(filename) {
    try {
        // read .csv file on a server
        const target = "data/" + filename + ".csv";

        const res = await fetch(target);

        return res.text();
    } catch (err) {
        console.log(err);
    }
}

async function readJson(filename) {
    try {
        return fetch("data/" + filename + ".json")
            .then((res) => res.json())
            .then((json) => {
                return json;
            });
    } catch (err) {
        console.log(err);
    }
}

export async function getCsvData() {
    let csv = await readCsv("candles");
    let csv_data = Papa.parse(csv, {
        dynamicTyping: true,
    }).data;

    // Remove last line in CSV-file (it is empty, which gives error to Highcharts)
    csv_data.pop();

    let json_data = await readJson("headers");

    return {
        candles: csv_data,
        headers: json_data,
    };
}
