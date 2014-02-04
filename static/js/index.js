$(document).ready(function(){
    connect_websocket();

    $("#url1-download-container").click(function() {
        var data = {
            "Target" : "url1",
            "Url"    : $("#url1").val()
        };
        ws.send(JSON.stringify(data));

        $(this).addClass("hide");
        $("#url1-wait-container").removeClass("hide");
        $("#url1-progress-container").removeClass("hide");
    })
})

function set_list_item_success($this) {
    $this.removeClass("list-group-item-info");
    $this.removeClass("list-group-item-danger");
    $this.addClass("list-group-item-success");
}

function set_list_item_error($this) {
    if ($this.hasClass("list-group-item-info")) {
        $this.removeClass("list-group-item-info");
        $this.addClass("list-group-item-danger");
    }
}

function connect_websocket() {
    ws = new WebSocket("ws://127.0.0.1:9090/download/");

    // First connect
    ws.onopen = function() {
        console.log("[onopen] connect ws uri.");
    }

    // Sending from server
    ws.onmessage = function(e) {
        var res = JSON.parse(e.data);
        if (res["Target"] == "url1") {
            if (res["Status"] == "ok") {
                setInterval(function(){
                    $("#url1-download-container").addClass("hide");
                    $("#url1-progress-container").addClass("hide");
                    $("#url1-status-ok").removeClass("hide");
                    $("#url1-status-fail").addClass("hide");
                    $("#url1-wait-container").addClass("hide");
                    $("#url1-play-container").removeClass("hide");
                    $("#url1").attr("disabled", "disabled");
                    set_list_item_success($("#url1-list-group-item"));
                }, 1000);
            } else if (res["Status"] == "keep") {
                if ($("#url1-play-container").hasClass("hide")) {
                    $("#url1-play-container").removeClass("hide");
                }
                $("#url1-progress-bar").css("width", res["Progress"]+"%");
            } else if (res["Status"] == "fail") {
                $("#url1-play-container").addClass("hide");
                $("#url1-wait-container").addClass("hide");
                $("#url1-download-container").removeClass("hide");
                $("#url1-progress-container").addClass("hide");
                $("#url1-status-ok").addClass("hide");
                $("#url1-status-fail").removeClass("hide");
                set_list_item_error($("#url1-list-group-item"));
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
