$(document).ready(function(){
    connect_websocket();

    $("#url1-play-container").click(function(){
        if ($("#ffmpeg-path").val() != "") {
            $.ajax({
                url: '/playVideo/',
                type: "POST",
                data: {FFmpegPath: $("#ffmpeg-path").val(), FilePath: $(this).attr("data-filepath")},
                Success: function(response) {
                    var res = JSON.parse(response);
                    if (res["Status"] == "fail") {
                        alert(res["ErrMsg"]);
                    }
                }
            })
        } else {
            alert("請選擇 ffmpeg 播放器路徑");
            $("#ffmpeg-path").focus();
        }
    })
    $("#url1-download-container").click(function() {
        var data = {
            "Target" : "url1",
            "Url"    : $("#url1").val()
        };
        ws.send(JSON.stringify(data));

        $(this).addClass("hide");
        $("#url1-wait-container").removeClass("hide");
        $("#url1-progress-container").removeClass("hide");
        set_list_item_warning($("#url1-list-group-item"))
    })
})

function set_list_item_success($this) {
    $this.removeClass("list-group-item-warning");
    $this.addClass("list-group-item-success");
}

function set_list_item_warning($this) {
    if ($this.hasClass("list-group-item-info")) {
        $this.removeClass("list-group-item-info");
        $this.addClass("list-group-item-warning");
    }
    if ($this.hasClass("list-group-item-danger")) {
        $this.removeClass("list-group-item-danger");
        $this.addClass("list-group-item-warning");
    }
}

function set_list_item_error($this) {
    $this.removeClass("list-group-item-warning");
    $this.addClass("list-group-item-danger");
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
                setTimeout(function(){
                    $("#url1-download-container").addClass("hide");
                    $("#url1-progress-container").addClass("hide");
                    $("#url1-status-ok").removeClass("hide");
                    $("#url1-status-fail").addClass("hide");
                    $("#url1-wait-container").addClass("hide");
                    $("#url1-play-container").removeClass("hide");
                    $("#url1").attr("disabled", "disabled");
                    set_list_item_success($("#url1-list-group-item"));
                    if ($("#url1-play-container").data("filepath") != "undefined") {
                        $("#url1-play-container").attr("data-filepath", res["FilePath"]);
                    }
                }, 1000);
            } else if (res["Status"] == "keep") {
                if ($("#url1-play-container").hasClass("hide")) {
                    $("#url1-play-container").attr("data-filepath", res["FilePath"]);
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