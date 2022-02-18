import {Container, Table, Tbody, Td, Text, Tr} from "@chakra-ui/react";

function FundsTable(props) {

    const transactions = props.data["transactions"];
    transactions.reverse();
    const infos = [];
    let allSameYear = true;
    let baseYear = stringToDate(transactions[0]["date"])["year"];
    transactions.forEach(tr => {
        let base = tr["base"];
        let fraction = tr["fraction"];
        let amount = base + ",";
        amount += (fraction === 0) ? "-" : fraction;
        let date = stringToDate(tr["date"])
        if (date["year"] !== baseYear) {
            allSameYear = false;
        }
        infos.push({"name": tr["name"], "amount": amount, "date": tr["date"]})
    });

    if(allSameYear) {
        infos.forEach(info => {
            const date = stringToDate(info["date"])
            info["date"] = "" + date["day"] + "." + date["month"] + "."
        })
    }

    function stringToDate(dateString) {
        const items = dateString.split(".")
        const day = items[0]
        const month = items[1]
        const year = items[2]
        return {"day": day, "month": month, "year": year}
    }

    return (
        <Container w={'90%'}>
            <Table variant='striped'>
                <Tbody>
                    {infos.map((item, idx) =>
                        <Tr key={"tr-" + idx}>
                            <Td key={"td1-" + idx}><Text key={"txt1-" + idx}>{item.date ? item.date : "20.02."}</Text></Td>
                            <Td key={"td2-" + idx}><Text key={"txt2-" + idx}>{item.name}</Text></Td>
                            <Td key={"td3-" + idx} isNumeric={true}><Text key={"txt3-" + idx}>{item.amount}</Text></Td>
                        </Tr>
                    )}
                </Tbody>
            </Table>
        </Container>
    )
}

export default FundsTable;