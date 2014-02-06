$(document).ready(function(){
    connect_websocket();

    $("#shutdown").click(function(){
        $.ajax({
            url: '/',
            type: "POST",
            data: {shutdown: 1}
        })
        location.reload();
    })
    $("[id^=url]").click(function(){
        var num = $(this).data("num");
        $("#url-" + num + "-play-container").click(function(){
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
        $("#url-" + num + "-download-container").click(function() {
            var data = {
                "Target" : "#url-" + num ,
                "Url"    : $("#url-" + num).val()
            };
            ws.send(JSON.stringify(data));

            $(this).addClass("hide");
            $("#url-" + num + "-wait-container").removeClass("hide");
            $("#url-" + num + "-progress-container").removeClass("hide");
            set_list_item_warning($("#url-" + num + "-list-group-item"))
        })
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
        if (res["Status"] == "ok") {
            setTimeout(function(){
                $(res["Target"] + "-download-container").addClass("hide");
                $(res["Target"] + "-progress-container").addClass("hide");
                $(res["Target"] + "-status-ok").removeClass("hide");
                $(res["Target"] + "-status-fail").addClass("hide");
                $(res["Target"] + "-wait-container").addClass("hide");
                $(res["Target"] + "-play-container").removeClass("hide");
                $(res["Target"] + "").attr("disabled", "disabled");
                set_list_item_success($(res["Target"] + "-list-group-item"));
                if ($(res["Target"] + "-play-container").data("filepath") != "undefined") {
                    $(res["Target"] + "-play-container").attr("data-filepath", res["FilePath"]);
                }
            }, 1000);
        } else if (res["Status"] == "keep") {
            if ($(res["Target"] + "-play-container").hasClass("hide")) {
                $(res["Target"] + "-play-container").attr("data-filepath", res["FilePath"]);
                $(res["Target"] + "-play-container").removeClass("hide");
            }
            $(res["Target"] + "-progress-bar").css("width", res["Progress"]+"%");
        } else if (res["Status"] == "fail") {
            $(res["Target"] + "-play-container").addClass("hide");
            $(res["Target"] + "-wait-container").addClass("hide");
            $(res["Target"] + "-download-container").removeClass("hide");
            $(res["Target"] + "-progress-container").addClass("hide");
            $(res["Target"] + "-status-ok").addClass("hide");
            $(res["Target"] + "-status-fail").removeClass("hide");
            set_list_item_error($(res["Target"] + "-list-group-item"));
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