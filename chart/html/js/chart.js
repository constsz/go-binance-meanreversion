// To use imports use "module":
// <script type="module" src="js/chart.js">
import { getCsvData } from "./readCsv.js";

// Get Data for OHLC, Volume, Indicators, etc
getCsvData().then((data) => {
    // Data from Go Backtester
    let candles = data.candles;
    let headers = data.headers;

    let sma_colors = ["#E21B9A", "#433BA7", "#6C3792", "#864b65", "#3E6CCA"];
    let bb_colors = ["#C6BEA6", "#FFBF00", "#987200", "#4E3A00"];

    // Create data-structure to hold indicator values
    /* 
    indicators = [
        {
            Name: "sma_fast",
            Type: "sma",
            Data: [
                
            ]
        }
    ]

    Далее, ниже, при создании Highchart, он лупом проходит по indicators,
    и если совпадает Type с его функцией, создает нужные индикаторы.
    
    */

    // Create a data-structures with all indicators
    let indicators = [];

    for (let i in headers) {
        let entry = headers[i];
        // If not a ohlcv-candle
        if (entry.Type != "ohlcv") {
            // Fill in details
            let indicator = {
                Name: entry.Name,
                Type: entry.Type,
                Data: [],
            };
            // Copy indicator Values to new array
            for (let j in candles) {
                // [timestamp, indicator_value]
                let indicator_value = [candles[j][0], candles[j][i]];
                indicator.Data.push(indicator_value);
            }
            indicators.push(indicator);
        }
    }

    console.log(indicators);

    // CREATE CHART
    const chart = Highcharts.stockChart(
        "container",
        {
            rangeSelector: {
                selected: 1,
            },
            series: [
                {
                    id: "A1",
                    name: "symbolName",
                    type: "candlestick",
                    data: candles,
                    yAxis: 0,
                    dataGrouping: {
                        enabled: false,
                    },
                    dashStyle: "Solid",
                },
            ],

            colors: ["#000000"],
            tooltip: {
                // crosshairs: false,
                crosshairs: {
                    color: "#666666",
                    dashStyle: "dash",
                },
                shared: true,
            },
            rangeSelector: {
                enabled: false,
            },
            navigation: {
                buttonOptions: {
                    enabled: false,
                },
            },

            chart: {
                backgroundColor: "#151923",
                style: {
                    fontFamily: "Roboto",
                    color: "#666666",
                },
            },
            title: {
                align: "left",
                style: {
                    fontFamily: "Roboto Condensed",
                    fontWeight: "bold",
                },
            },
            subtitle: {
                align: "left",
                style: {
                    fontFamily: "Roboto Condensed",
                },
            },
            legend: {
                align: "right",
                verticalAlign: "bottom",
                itemStyle: {
                    color: "#424242",
                },
            },
            xAxis: {
                gridLineColor: "#151923",
                gridLineWidth: 1,
                minorGridLineColor: "#424242",
                minoGridLineWidth: 0.5,
                tickColor: "#fffff",
                minorTickColor: "#424242",
                lineColor: "#424242",
            },
            yAxis: {
                gridLineColor: "#151923",
                ridLineWidth: 1,
                minorGridLineColor: "#424242",
                inoGridLineWidth: 0.5,
                tickColor: "#424242",
                minorTickColor: "#424242",
                lineColor: "#424242",
                height: "100%",
            },
        },
        (chart) => {
            // CREATE INDICATORS FOR HIGHCHARTS
            let sma_counter = 0;
            for (let i in indicators) {
                // Choose the Type of Indicator to be created
                switch (indicators[i].Type) {
                    case "SMA":
                        chart.addSeries({
                            name: indicators[i].Name,
                            id: indicators[i].Name,
                            data: indicators[i].Data,
                            color: sma_colors[sma_counter],
                        });
                        sma_counter += 1;
                        break;

                    case "TrendMove_TrendLine":
                        chart.addSeries({
                            name: indicators[i].Name,
                            id: indicators[i].Name,
                            data: indicators[i].Data,
                            color: sma_colors[0],
                        });
                        break;

                    // case "TrendMove_MoveLine":
                    //     chart.addSeries({
                    //         name: indicators[i].Name,
                    //         id: indicators[i].Name,
                    //         data: indicators[i].Data,
                    //         color: sma_colors[2],
                    //     });
                    //     break;

                    case "BetterBands_LowBand":
                        chart.addSeries({
                            name: indicators[i].Name,
                            id: indicators[i].Name,
                            data: indicators[i].Data,
                            color: bb_colors[2],
                        });
                        break;
                    case "BetterBands_HighBand":
                        chart.addSeries({
                            name: indicators[i].Name,
                            id: indicators[i].Name,
                            data: indicators[i].Data,
                            color: bb_colors[2],
                        });
                        break;

                    // case "ATR":
                    //     chart.addAxis({
                    //         id: "axis-" + indicators[i].Name,
                    //         title: { text: "ATR" },
                    //         lineWidth: 0,
                    //         gridLineColor: "#151923",
                    //         ridLineWidth: 1,
                    //         minorGridLineColor: "#424242",
                    //         inoGridLineWidth: 0.5,
                    //         tickColor: "#424242",
                    //         minorTickColor: "#424242",
                    //         lineColor: "#424242",
                    //         height: "20%",
                    //         top: "80%",
                    //     });
                    //     chart.addSeries({
                    //         name: indicators[i].Name,
                    //         yAxis: "axis-" + indicators[i].Name,
                    //         id: indicators[i].Name,
                    //         data: indicators[i].Data,
                    //         color: "#3ECABD",
                    //     });
                    //     break;
                    //     case "superTrend_value":
                    //         // case "superTrend_direction":
                    //         chart.addSeries({
                    //             name: indicators[i].Name,
                    //             id: indicators[i].Name,
                    //             data: indicators[i].Data,
                    //             color: "#3ECABD",
                    //         });
                    //         break;
                    //     case "smi_smi_signal":
                    //         chart.addAxis({
                    //             id: "axis-" + indicators[i].Name,
                    //             title: { text: "smi" },
                    //             lineWidth: 0,
                    //             gridLineColor: "#151923",
                    //             ridLineWidth: 1,
                    //             minorGridLineColor: "#424242",
                    //             inoGridLineWidth: 0.5,
                    //             tickColor: "#424242",
                    //             minorTickColor: "#424242",
                    //             lineColor: "#424242",
                    //             height: "20%",
                    //             top: "80%",
                    //         });
                    //         chart.addSeries({
                    //             name: indicators[i].Name,
                    //             yAxis: "axis-" + indicators[i].Name,
                    //             id: indicators[i].Name,
                    //             data: indicators[i].Data,
                    //             color: "#33D1FF",
                    //         });
                    //         break;
                    //     case "smi_ema_signal":
                    //         chart.addSeries({
                    //             name: indicators[i].Name,
                    //             yAxis: "axis-" + indicators[i].Name,
                    //             id: indicators[i].Name,
                    //             data: indicators[i].Data,
                    //             color: "#FFCA30",
                    //         });
                }
            }

            // old code - just saved
            // "SOME REMOVAL"
            // let series = chart.get("A1");
            // series.remove();
            // chart.addSeries({
            //     id: "B2",
            //     data: data,
            // });

            // "ENTRY LONG"
            // if (entry_long.length > 0) {
            //     chart.addSeries({
            //         name: "Entry Long",
            //         type: "scatter",
            //         data: entry_long,
            //         marker: {
            //             symbol: "url(./icons/green_triangle_up.png)",
            //             width: 32,
            //             height: 32,
            //         },
            //         dataLabels: {
            //             enabled: false,
            //             color: "#fff",
            //             y: 12,
            //         },
            //     });
            // }
        }
    );
});
