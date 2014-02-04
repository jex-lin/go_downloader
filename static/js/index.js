$(document).ready(function(){
    connect_websocket();
    $("#url1").change(function() {
        $("#url1-btn").removeAttr("disabled");
        $("#url1-btn").removeClass("btn-danger btn-success").addClass("btn-default");
        $("#url1-btn").html("Download");
        $("#url1-progress").show();
        $("#url1-show").css("width", "1%");
    })
    $("#url1-btn").click(function() {
        var data = {
            "Target" : "url1",
            "Url"    : $("#url1").val()
        };
        ws.send(JSON.stringify(data));
        $("#url1-btn").removeClass("btn-default btn-danger btn-success").addClass("btn-warning");
        $("#url1-btn").html("Downloading...")
        $("#url1-btn").attr("disabled", "disabled");
        $("#url1-progress").show();
    })
})

function connect_websocket() {
    ws = new WebSocket("ws://127.0.0.1:9090/download/");

    // First connect
    ws.onopen = function() {
        console.log("[onopen] connect ws uri.");
    }

    // Sending from server
    ws.onmessage = function(e) {
        var res = JSON.parse(e.data);
        console.log(res);
        if (res["Target"] == "url1") {
            if (res["Status"] == "ok") {
                setInterval(function(){
                    $("#url1-btn").removeClass("btn-warning");
                    $("#url1-btn").addClass("btn-success");
                    $("#url1-btn").html("Success");
                    $("#url1-progress").hide();
                },1000);
            } else if (res["Status"] == "keep") {
                $("#url1-show").css("width", res["Progress"]+"%");
            } else if (res["Status"] == "fail") {
                $("#url1-progress").hide();
                $("#url1-btn").removeClass("btn-warning");
                $("#url1-btn").removeAttr("disabled");
                $("#url1-btn").addClass("btn-danger");
                $("#url1-btn").html("Retry");
            }
        }
    }

    // Server close connection
    ws.onclose = function(e) {
        console.log("[onclose] connection closed (" + e.code + ")");
    }

    // Occur error
    ws.onerror = function (e) {
        console.log("[onerror] error!");
    }
}
//$.ajax({
//        url: '/api/',
//        type: "POST",
//        data: {target: "url1", url: $("#url1").val()},
//        beforeSend: function() {
//            $("#url1-btn").removeClass("btn-default btn-danger btn-success").addClass("btn-warning");
//            $("#url1-btn").html("Downloading...")
//            $("#url1-btn").attr("disabled", "disabled");
//            $("#url1-progress").show();
//        },
//        success: function(response) {
//            $("#url1-btn").removeClass("btn-warning");
//            var res = JSON.parse(response);
//            if (res["status"] == "ok") {
//                $("#url1-btn").addClass("btn-success");
//                $("#url1-btn").html("Success");
//            } else {
//                $("#url1-btn").removeAttr("disabled");
//                $("#url1-btn").addClass("btn-danger");
//                $("#url1-btn").html("Retry");
//            }
//            $("#url1-progress").hide();
//        },
//})
