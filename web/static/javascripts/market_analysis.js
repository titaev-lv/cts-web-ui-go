$(document).ready(function() {
    $('#exchange_1').on('change', function(e) {
        var ex = $('#exchange_1 option:selected').val();
        if(ex !== '') {
            $.ajax({
                url: "/market_analysis/ajax_exchanges_step_1.php",
                type: "POST", 
                dataType: "html",
                data: {'exchange_id':ex},
                /*beforeSend: function(xhr) {
                    xhr.setRequestHeader("Content-type", "multipart/form-data");
                },*/
                success: function(response) { //Данные отправлены успешно
                    var ret = JSON.parse(response);
                    if(ret.error !== false && ret.error !== '') {
                        new PNotify({
                                title: 'Error',
                                text: ret.error,
                                type: 'error',
                                addclass: 'stack-bar-top',

                                width: "100%"
                        });
                    }
                    else if(ret.success === true) {
                         new PNotify({
                                /*title: 'OK',*/
                                text: 'Trading pairs are synchronized with the exchange',
                                type: 'success',
                                addclass: 'stack-bar-top',
                                width: "100%"
                        });
                        if(ret.data !== false && ret.data !== '') {
                            var d = ret.data;
                            $('#exchange_2 option').each(function() {
                                $(this).remove();
                            });
                            $('#exchange_2').append('<option value=""></option>');
                            $('#trade_pair').val('');
                            for(var i in d) {
                                $('#exchange_2').append('<option value="'+d[i]['ID']+'">'+d[i]['NAME']+'</option>');
                            }
                        }                   
                    }
                },
                error: function (data, textStatus) {
                    if(data.status == 401) {
                        setTimeout(function(){ location.reload(); }, 1000);
                    }
                    new PNotify({
                                title: 'Error',
                                text: "Error " + data.status + " " + data.statusText,
                                type: 'error',
                                addclass: 'stack-bar-top',
                                width: "100%"
                        });
                }  
            });
            $('#fee_1_maker').text('');
            $('#fee_1_taker').text('');
            $('#fee_2_maker').text('');
            $('#fee_2_taker').text('');
        }
        else {
            $('#exchange_2 option').each(function() {
                $(this).remove();
            });
            $('#trade_pair').val('');
            $('#trade_list option').each(function() {
                $(this).remove();
            }); 
            $('#fee_1_maker').text('');
            $('#fee_1_taker').text('');
            $('#fee_2_maker').text('');
            $('#fee_2_taker').text('');
        }
    });
    $('#exchange_2').on('change', function(e) {
        var ex1 = $('#exchange_1 option:selected').val();
        var ex2 = $('#exchange_2 option:selected').val();
        $('#trade_list').append('<option value=""></option>');
        $('#trade_pair').val('');
        if(ex1 !== '' && ex2 !== '') {
           $.ajax({
                url: "/market_analysis/ajax_exchanges_step_2.php",
                type: "POST", 
                dataType: "html",
                data: {'exchange_id1':ex1,'exchange_id2':ex2},
                /*beforeSend: function(xhr) {
                    xhr.setRequestHeader("Content-type", "multipart/form-data");
                },*/
                success: function(response) { //Данные отправлены успешно
                    var ret = JSON.parse(response);
                    if(ret.error !== false && ret.error !== '') {
                        new PNotify({
                                title: 'Error',
                                text: ret.error,
                                type: 'error',
                                addclass: 'stack-bar-top',

                                width: "100%"
                        });
                    }
                    else if(ret.success === true) {
                         new PNotify({
                                /*title: 'OK',*/
                                text: 'Trading pairs are synchronized with the exchange',
                                type: 'success',
                                addclass: 'stack-bar-top',
                                width: "100%"
                        });
                        if(ret.data !== false && ret.data !== '') {
                            var d = ret.data;
                            
                            $('#trade_list option').each(function() {
                                $(this).remove();
                            });                            
                            for(var i in d) {
                                $('#trade_list').append('<option value="'+d[i]['pair']+'">'+d[i]['pair']);
                            }
                        }                   
                    }
                },
                error: function (data, textStatus) {
                    if(data.status == 401) {
                        setTimeout(function(){ location.reload(); }, 1000);
                    }
                    new PNotify({
                                title: 'Error',
                                text: "Error " + data.status + " " + data.statusText,
                                type: 'error',
                                addclass: 'stack-bar-top',
                                width: "100%"
                        });
                }  
            });
            $('#fee_1_maker').text('');
            $('#fee_1_taker').text('');
            $('#fee_2_maker').text('');
            $('#fee_2_taker').text('');
        }
        else {
            $('#trade_list option').each(function() {
                $(this).remove();
            }); 
            $('#fee_1_maker').text('');
            $('#fee_1_taker').text('');
            $('#fee_2_maker').text('');
            $('#fee_2_taker').text('');
        }
    });

    $('#trade_pair').on('input', function() {
        var ex1 = $('#exchange_1 option:selected').val();
        var ex2 = $('#exchange_2 option:selected').val();
        var userText = $(this).val();
        $("#trade_list").find("option").each(function () {
            if ($(this).val() == userText) {
                $.ajax({
                    url: "/market_analysis/ajax_exchanges_step_3.php",
                    type: "POST", 
                    dataType: "html",
                    data: {'exchange_id1':ex1,'exchange_id2':ex2, 'trade_pair': userText},
                    /*beforeSend: function(xhr) {
                        xhr.setRequestHeader("Content-type", "multipart/form-data");
                    },*/
                    success: function(response) { //Данные отправлены успешно
                        var ret = JSON.parse(response);
                        if(ret.error !== false && ret.error !== '') {
                            new PNotify({
                                    title: 'Error',
                                    text: ret.error,
                                    type: 'error',
                                    addclass: 'stack-bar-top',

                                    width: "100%"
                            });
                        }
                        else if(ret.success === true) {
                             new PNotify({
                                    /*title: 'OK',*/
                                    text: 'Trading Fee are synchronized with the exchanges',
                                    type: 'success',
                                    addclass: 'stack-bar-top',
                                    width: "100%"
                            });
                            if(ret.data !== false && ret.data !== '') {
                                var d = ret.data;

                                $('#timeframe option').each(function() {
                                    $(this).remove();
                                }); 
                                $('#timeframe').append('<option value="">');
                                for(var i in d) {
                                    $('#timeframe').append('<option value="'+d[i]+'">'+d[i]);
                                }
                            }       
                            
                            if(ret.fee !== false){
                                 $('#fee_1_maker').text(ret.fee.ex1.maker_fee*100+'%');
                                 $('#fee_1_taker').text(ret.fee.ex1.taker_fee*100+'%');
                                 $('#fee_2_maker').text(ret.fee.ex2.maker_fee*100+'%');
                                 $('#fee_2_taker').text(ret.fee.ex2.taker_fee*100+'%');
                            }
                            am5.array.each(am5.registry.rootElements, function(root) {
                                root.dispose();
                            });
                        }
                    },
                    error: function (data, textStatus) {
                        if(data.status == 401) {
                            setTimeout(function(){ location.reload(); }, 1000);
                        }
                        new PNotify({
                                    title: 'Error',
                                    text: "Error " + data.status + " " + data.statusText,
                                    type: 'error',
                                    addclass: 'stack-bar-top',
                                    width: "100%"
                            });
                    }  
                });
            }
        });
    });
    
    $('#timeframe').on('input', function() {
        var ex1 = $('#exchange_1 option:selected').val();
        var ex2 = $('#exchange_2 option:selected').val();
        var trade_pair = $('#trade_pair').val();
        var timeframe = $('#timeframe option:selected').val();
        $.ajax({
            url: "/market_analysis/ajax_exchanges_step_4.php",
            type: "POST", 
            dataType: "html",
            data: {'exchange_id1':ex1,'exchange_id2':ex2, 'trade_pair': trade_pair, 'timeframe':timeframe},
            success: function(response) { //Данные отправлены успешно
                var ret = JSON.parse(response);
                if(ret.error !== false && ret.error !== '') {
                    new PNotify({
                            title: 'Error',
                            text: ret.error,
                            type: 'error',
                            addclass: 'stack-bar-top',

                            width: "100%"
                    });
                }
                else if(ret.success === true) {
                     new PNotify({
                            /*title: 'OK',*/
                            text: 'Trading Fee are synchronized with the exchanges',
                            type: 'success',
                            addclass: 'stack-bar-top',
                            width: "100%"
                    });
                    
                    var data = ret.data;
                    
                    //am5.ready(function() {
                        // Create root element
                        // https://www.amcharts.com/docs/v5/getting-started/#Root_element
                       
                        am5.array.each(am5.registry.rootElements, function(root) {
                             root.dispose();
                        });
                       /* var root = am5.Root.new("chartdiv");

                        //root.container.children.clear();

                        // Set themes
                        // https://www.amcharts.com/docs/v5/concepts/themes/
                        root.setThemes([
                          am5themes_Animated.new(root)
                        ]);
                        // Create chart
                        // https://www.amcharts.com/docs/v5/charts/xy-chart/
                        var chart = root.container.children.push(
                            am5xy.XYChart.new(root, {
                                panX: true,
                                panY: true,
                                wheelX: "panX",
                                wheelY: "zoomX",
                                pinchZoomX:true
                            })
                        );
                        chart.get("colors").set("step", 5);
                        // Add cursor
                        // https://www.amcharts.com/docs/v5/charts/xy-chart/cursor/
                        var cursor = chart.set(
                            "cursor",
                            am5xy.XYCursor.new(root, {
                                behavior: "none"
                            })
                        );
                        cursor.lineY.set("visible", false);
                        // Create axes
                        // https://www.amcharts.com/docs/v5/charts/xy-chart/axes/
                        */
                        switch(timeframe) {
                            case '1min':
                                var t1 = 'minute';
                                var t2 = 1;
                                break;
                             case '3min':
                                var t1 = 'minute';
                                var t2 = 3;
                                break;
                            case '5min':
                                var t1 = 'minute';
                                var t2 = 5;
                                break;
                            case '10min':
                                var t1 = 'minute';
                                var t2 = 10;                               
                                break;
                            case '15min':
                                var t1 = 'minute';
                                var t2 = 15;                                
                                break;
                            case '30min':
                                var t1 = 'minute';
                                var t2 = 30;                                
                                break;
                            case '1hour':
                                var t1 = 'hour';
                                var t2 = 1;                               
                                break;
                            case '2hour':
                                var t1 = 'hour';
                                var t2 = 2;                               
                                break;
                            case '4hour':
                                var t1 = 'hour';
                                var t2 = 4;                               
                                break;
                        }
                        
                       /* var xAxis = chart.xAxes.push(
                            am5xy.DateAxis.new(root, {
                                baseInterval: { timeUnit: t1, count: t2 },
                                renderer: am5xy.AxisRendererX.new(root, {}),
                                tooltip: am5.Tooltip.new(root, {})
                            })
                        );
                        var yAxis = chart.yAxes.push(
                            am5xy.ValueAxis.new(root, {
                                renderer: am5xy.AxisRendererY.new(root, {})
                            })
                        );
                        // Add series
                        // https://www.amcharts.com/docs/v5/charts/xy-chart/series/
                        var series1 = chart.series.push(
                            am5xy.LineSeries.new(root, {
                                name: "Series",
                                xAxis: xAxis,
                                yAxis: yAxis,
                                valueYField: "open",
                                openValueYField: "close",
                                valueXField: "date",
                                stroke: root.interfaceColors.get("positive"),
                                fill: root.interfaceColors.get("positive"),
                                tooltip: am5.Tooltip.new(root, {
                                    labelText: "{valueY.formatNumber('#.00000000')}"
                                }),
                            })
                        );

                        series1.fills.template.setAll({
                          fillOpacity: 0.6,
                          visible: true
                        });

                        var series2 = chart.series.push(
                            am5xy.LineSeries.new(root, {
                                name: "Series",
                                xAxis: xAxis,
                                yAxis: yAxis,
                                valueYField: "close",
                                valueXField: "date",
                                stroke: root.interfaceColors.get("negative"),
                                fill: root.interfaceColors.get("negative"),
                                tooltip: am5.Tooltip.new(root, {
                                    labelText: "{valueY.formatNumber('#.00000000')}"
                                })
                            })
                        );
                        // Add scrollbar
                        // https://www.amcharts.com/docs/v5/charts/xy-chart/scrollbars/
                        chart.set("scrollbarX", am5.Scrollbar.new(root, {
                            orientation: "horizontal"
                        }));

                        //var data = ;
                        series1.data.setAll(data);
                        series2.data.setAll(data);

                        // create ranges
                        var i = 0;
                        var baseInterval = xAxis.get("baseInterval");
                        var baseDuration = xAxis.baseDuration();
                        var rangeDataItem;

                        am5.array.each(series1.dataItems, function (s1DataItem) {
                          var s1PreviousDataItem;
                          var s2PreviousDataItem;

                          var s2DataItem = series2.dataItems[i];

                          if (i > 0) {
                            s1PreviousDataItem = series1.dataItems[i - 1];
                            s2PreviousDataItem = series2.dataItems[i - 1];
                          }

                          var startTime = am5.time
                            .round(
                              new Date(s1DataItem.get("valueX")),
                              baseInterval.timeUnit,
                              baseInterval.count
                            )
                            .getTime();

                          // intersections
                          if (s1PreviousDataItem && s2PreviousDataItem) {
                            var x0 =
                              am5.time
                                .round(
                                  new Date(s1PreviousDataItem.get("valueX")),
                                  baseInterval.timeUnit,
                                  baseInterval.count
                                )
                                .getTime() +
                              baseDuration / 2;
                            var y01 = s1PreviousDataItem.get("valueY");
                            var y02 = s2PreviousDataItem.get("valueY");

                            var x1 = startTime + baseDuration / 2;
                            var y11 = s1DataItem.get("valueY");
                            var y12 = s2DataItem.get("valueY");

                            var intersection = getLineIntersection(
                              { x: x0, y: y01 },
                              { x: x1, y: y11 },
                              { x: x0, y: y02 },
                              { x: x1, y: y12 }
                            );

                            startTime = Math.round(intersection.x);
                        }

                          // start range here
                          if (s2DataItem.get("valueY") > s1DataItem.get("valueY")) {
                            if (!rangeDataItem) {
                              rangeDataItem = xAxis.makeDataItem({});
                              var range = series1.createAxisRange(rangeDataItem);
                              rangeDataItem.set("value", startTime);
                              range.fills.template.setAll({
                                fill: series2.get("fill"),
                                fillOpacity: 0.6,
                                visible: true
                              });
                              range.strokes.template.setAll({
                                stroke: series1.get("stroke"),
                                strokeWidth: 1
                              });
                            }
                          } else {
                            // if negative range started
                            if (rangeDataItem) {
                              rangeDataItem.set("endValue", startTime);
                            }

                            rangeDataItem = undefined;
                          }
                          // end if last
                          if (i == series1.dataItems.length - 1) {
                            if (rangeDataItem) {
                              rangeDataItem.set(
                                "endValue",
                                s1DataItem.get("valueX") + baseDuration / 2
                              );
                              rangeDataItem = undefined;
                            }
                          }

                          i++;
                        });

                        // Make stuff animate on load
                        // https://www.amcharts.com/docs/v5/concepts/animations/
                        series1.appear(1000);
                        series2.appear(1000);
                        chart.appear(1000, 100);

                        function getLineIntersection(pointA1, pointA2, pointB1, pointB2) {
                          let x =
                            ((pointA1.x * pointA2.y - pointA2.x * pointA1.y) * (pointB1.x - pointB2.x) -
                              (pointA1.x - pointA2.x) *
                                (pointB1.x * pointB2.y - pointB1.y * pointB2.x)) /
                            ((pointA1.x - pointA2.x) * (pointB1.y - pointB2.y) -
                              (pointA1.y - pointA2.y) * (pointB1.x - pointB2.x));
                          let y =
                            ((pointA1.x * pointA2.y - pointA2.x * pointA1.y) * (pointB1.y - pointB2.y) -
                              (pointA1.y - pointA2.y) *
                                (pointB1.x * pointB2.y - pointB1.y * pointB2.x)) /
                            ((pointA1.x - pointA2.x) * (pointB1.y - pointB2.y) -
                              (pointA1.y - pointA2.y) * (pointB1.x - pointB2.x));
                          return { x: x, y: y };
                        }
                        */
                        ///////////////////////////////////////
                        //////////////////////////////////////
                        
                        var root2 = am5.Root.new("ch2");
                        root2.setThemes([
                            am5themes_Animated.new(root2)
                        ]);
                        // Create chart
                        // https://www.amcharts.com/docs/v5/charts/xy-chart/
                        var chart2 = root2.container.children.push(am5xy.XYChart.new(root2, {
                            panX: true,
                            panY: true,
                            wheelX: "panX",
                            wheelY: "zoomX",
                            pinchZoomX:true
                        }));
                        // Add cursor
                        // https://www.amcharts.com/docs/v5/charts/xy-chart/cursor/
                        var cursor2 = chart2.set("cursor", am5xy.XYCursor.new(root2, {
                            behavior: "none"
                        }));
                        cursor2.lineY.set("visible", false);

                        // Create axes
                        // https://www.amcharts.com/docs/v5/charts/xy-chart/axes/
                        var xAxis = chart2.xAxes.push(am5xy.DateAxis.new(root2, {
                            //maxDeviation: 0.2,
                            baseInterval: {
                                timeUnit: t1,
                                count: t2
                            },
                            renderer: am5xy.AxisRendererX.new(root2, {}),
                            tooltip: am5.Tooltip.new(root2, {})
                        }));

                        var yAxis = chart2.yAxes.push(am5xy.ValueAxis.new(root2, {
                            numberFormat: "#.000000000",
                            strictMinMax: true,
                            renderer: am5xy.AxisRendererY.new(root2, {})
                        }));
                        // Add series
                        // https://www.amcharts.com/docs/v5/charts/xy-chart/series/
                        var series = chart2.series.push(am5xy.LineSeries.new(root2, {
                            name: "Series",
                            xAxis: xAxis,
                            yAxis: yAxis,
                            valueYField: "profit",
                            valueXField: "date",
                            tooltip: am5.Tooltip.new(root2, {
                                labelText: "{valueY.formatNumber('#.000000000')}"
                            })
                        }));
                        // Add scrollbar
                        // https://www.amcharts.com/docs/v5/charts/xy-chart/scrollbars/
                        chart2.set("scrollbarX", am5.Scrollbar.new(root2, {
                          orientation: "horizontal"
                        }));

                        // Set data
                        var data = ret.data2;
                        series.data.setAll(data);

                        // Make stuff animate on load
                        // https://www.amcharts.com/docs/v5/concepts/animations/
                        series.appear(1000);
                        chart2.appear(1000, 100);

                }
            },
            error: function (data, textStatus) {
                if(data.status == 401) {
                    setTimeout(function(){ location.reload(); }, 1000);
                }
                new PNotify({
                            title: 'Error',
                            text: "Error " + data.status + " " + data.statusText,
                            type: 'error',
                            addclass: 'stack-bar-top',
                            width: "100%"
                    });
            }  
        });
    });
    
    
    // Our labels and three data series
    /*var data = {
      labels: ['Week1', 'Week2', 'Week3', 'Week4', 'Week5', 'Week6'],
      series: [
        [5, 4, 3, 7, 5, 10],
        [3, 2, 9, 5, 4, 6],
        [2, 1, -3, -4, -2, 0]
      ]
    };

    // We are setting a few options for our chart and override the defaults
    var options = {
      // Don't draw the line chart points
      showPoint: false,
      // Disable line smoothing
      lineSmooth: false,
      // X-Axis specific configuration
      axisX: {
        // We can disable the grid for this axis
        showGrid: false,
        // and also don't show the label
        showLabel: false
      },
      // Y-Axis specific configuration
      axisY: {
        // Lets offset the chart a bit from the labels
        offset: 60,
        // The label interpolation function enables you to modify the values
        // used for the labels on each axis. Here we are converting the
        // values into million pound.
        labelInterpolationFnc: function(value) {
          return '$' + value + 'm';
        }
      }
    };

    // All you need to do is pass your configuration as third parameter to the chart function
    new Chartist.Line('.ct-chart', data, options);
    */

    $('#get_direct_ex_arb_button').on('click', function() {
        var symbol = $('#symbol option:selected').val();
        var start = $('#date_start').val();
        var stop = $('#date_stop').val();
        var error = false;
        var msg = '';
        var pattern = /\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}/g;
        
        //alert(symbol+start+stop);
        
        if(!symbol) {
            error = true;
            msg = "Symbol field is required";
            $('#symbol').parent().parent().addClass("has-error");
        }
        else {
            $('#symbol').parent().parent().removeClass("has-error");
        }
               
        if(!start.match(pattern)) {
            $('#date_start').parent().addClass("has-error");
            error = true;
            msg = "Symbol field is required";
        }
        else {
            $('#date_start').parent().removeClass("has-error");
        }
        if(!stop.match(pattern)) {
            $('#date_stop').parent().addClass("has-error");
            error = true;
            msg = "Symbol field is required";
        }
        else {
            $('#date_stop').parent().removeClass("has-error");
        }
        
        if(!error) {
            var date_start = new Date(start);
            var unixTimeStamp_start = Math.floor(date_start.getTime() / 1000);
            var date_stop = new Date(stop);
            var unixTimeStamp_stop = Math.floor(date_stop.getTime() / 1000);
            if(unixTimeStamp_start > unixTimeStamp_stop) {
                error = true;
                msg = "Time Interval failed set";
            }
        }
        if(!error) {
            //Reset table filter
            var h = table.columns().header();
            h.toArray().forEach(function(item, i, arr) {
                if(item.children[0]) {
                    item.children[0].value = '';
                    var id = item.children[0].id.split('row_');
                    id = id[1];
                    table.columns(id).search('');
                }
            });
            table.draw();
            /*$.ajax({
                url: "/market_analysis/ajax_direct_exs.php",
                type: "POST", 
                dataType: "html",
                data: {'symbol':symbol,'date_start':start, 'date_stop': stop},
                success: function(response) {
                    alert(1);
                },
                error: function (data, textStatus) {
                    
                }
            });   */  
        }
        
        if(error === true) {
            new PNotify({
                    title: 'Error',
                    text: "Error " + msg,
                    type: 'error',
                    addclass: 'stack-bar-top',
                    width: "100%"
            });
        }
    });
    
    var dtmaexs = document.getElementById('dt-direct_exs');
    if(dtmaexs) {
        //Insert into header table input field for search
        var nm = Array(
                "",
                "exs_time", 
                "exs_sell",
                "exm_buy",  
                "exm_price_sell",
                "exm_price_buy",
                "exm_fees",
                "exm_max_volume",
                "exm_profit",
                "exm_profit_total",
             );
        $('#dt-direct_exs thead tr th').each(function (i) {
            var title = $(this).text();
            $(this).html(title+' <input type="text" name="'+nm[i]+'@'+i+'" class="form-control input-sm mb-md input-search" placeholder="" style="padding:1px" onclick="event.stopPropagation();" onkeypress="event.stopPropagation();keysearchUser(event)" />');
        });
        	
        $.fn.dataTable.ext.errMode = 'throw';
        
        table =  $('#dt-direct_exs').DataTable( {
            "processing": true,
            "serverSide": true,
            "searching": true,
            "ordering": false,
            "scrollX": true,
            "pageLength": 50,
            "lengthMenu": [50, 100, 200, 500 ],
            "paging": true,
            //"pagingType": "simple",
            //"pagingType": "first_last_numbers",
            "pagingExtraNumberForNext": true,
           /* "bAutoWidth": false,*/
            "bScrollCollapse": true,
         //   "rowId": 'id',
            "columns": [
                { "data": "time" },              //0             
                { "data": "sell"},            //1
                { "data": "buy"},           //2
                { "data": "price_sell" },       //3
                { "data": "price_buy" },       //4
                { "data": "fees" },       //5
                { "data": "volume_max" },       //6
                { "data": "profit" },       //7
                { "data": "profit_total" }       //8
            ],
            "language": {
                 "processing": "Processing...",
                 "lengthMenu": "_MENU_ exchanges per page",
                 "zeroRecords": "Data not found",
                 "info": "Filtered from _START_ to _END_ of _TOTAL_",
                 "infoEmpty": "",
                 "infoFiltered": "(Total exchanges _MAX_)"
            },
            "ajax": {
                "method": "POST",
                "url": "/market_analysis/ajax_direct_exs.php",
                "data": function ( d ) {
                    d.date_start = $('#date_start').val();
                    d.date_stop = $('#date_stop').val();
                    d.symbol = $('#symbol option:selected').val();
                },
                "error": function (xhr, error, thrown) {
                },
                "statusCode": {
                    401: function (xhr, error, thrown) {
                        location.reload();
                    }
                }
            },
            "createdRow":function  ( row, data, dataIndex, cells ) {
            },
            "columnDefs": [
                    {
                        "targets": "_all",
                        "className": 'dt-body-left',
                        "searchable": true
                    },
            ],
            "select": {
                "style":    'os',
                "selector": 'td:first-child'
            },
            "order": [[ 1, 'desc' ]],
            /*"drawCallback": function( settings ) {
                console.log(settings);
            },  */
            "fnDrawCallback": function( oSettings ) {
                var ret_json = oSettings.json;
                if(ret_json.errors) {
                    new PNotify({
                        title: 'Error',
                        text: ret_json.errors,
                        type: 'error',
                        addclass: 'stack-bar-top',
                        width: "100%"
                    });
                }
                //console.log(ret_json.a);//do whatever with your custom response
            },
            "rowsGroup": [0]
        } );
         //Hide field search
        var search = document.getElementById('dt-direct_exs_filter');
        if (search) {
              document.getElementById('dt-direct_exs_filter').style.display = 'none';
        }
    }

 });
