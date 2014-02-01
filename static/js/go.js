$(document).ready(function(){
    $("#url1").change(function() {
        $("#dow1").removeAttr("disabled");
        $("#dow1").removeClass("btn-danger btn-success").addClass("btn-default");
        $("#dow1").html("Download");
    })
    $("#dow1").click(function() {
        $.ajax({
                url: '/api/',        //指向你要請求的PHP
                type: "POST",                        //如果要使用GET, 就改成 type: "GET",
                data: {url1: $("#url1").val()},                //或是用這種寫法 data: {test:1, test2:33},
                beforeSend: function() {
                    $("#dow1").removeClass("btn-default btn-danger btn-success").addClass("btn-warning");
                    $("#dow1").html("Downloading...")
                    $("#dow1").attr("disabled", "disabled");
                },
                success: function(response) {
                    $("#dow1").removeClass("btn-warning");
                    var res = JSON.parse(response);
                    if (res["status"] == "ok") {
                        $("#dow1").addClass("btn-success");
                        $("#dow1").html("Success");
                    } else {
                        $("#dow1").removeAttr("disabled");
                        $("#dow1").addClass("btn-danger");
                        $("#dow1").html("Retry");
                    }
                },
        })
    })
})
