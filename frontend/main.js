let endpoint = "https://{ YOUR API DOMAIN }/getDetails/";
let apiKey = "{ YOUR API KEY }";
let paypalLink = "{ YOUR PAYPAL LINK }";

window.onload = function() {
    let mpName = "";

    document.getElementById("paypal-link").href = paypalLink;

    let params = window.location.href.split("?")[1].split("&");
    for(let i = 0; i < params.length; i++) {
        let param = params[i];
        let kv = param.split("=");
        if (kv[0].toLowerCase() === "mp") {
            mpName = kv[1];
        }
    }

    fetch(endpoint + mpName, {
        "headers": {
            "x-api-key": apiKey
        }
    })
      .then(response => response.json())
      .then(data => {
          let transactions = data["transactions"];
          let name = data["name"];
          let title = data["title"];
          document.getElementById("subject").innerText = name;
          document.getElementById("moneypool-title").innerText = title;
          let sum = 0;
          transactions.forEach(tr => {
              let base = tr["base"];
              let fraction = tr["fraction"];
              sum += base + (fraction/100);
              let amount = base  + ",";
              amount += (fraction === 0) ? "-" : fraction;
              row(tr["name"], amount + "€");

              document.getElementById("sum").innerText = parseFloat(sum).toFixed(2) + "€";
          });
      });

};

function row(name, amount) {
    let table = document.getElementById("transactions-table");
    console.log(table);
    let tr = document.createElement("tr");
    td(name, "name", tr);
    td("contributed", "contrib-text", tr);
    td(amount, "amount", tr);
    table.appendChild(tr)

}

function td(text, className, parent) {
    let element = document.createElement("td");
    element.classList += className;
    element.innerText = text;
    parent.appendChild(element);
    return element;
}