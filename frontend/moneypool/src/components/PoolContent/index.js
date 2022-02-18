import {Center, Container, Skeleton, Stack} from "@chakra-ui/react";
import InfoSection from "./InfoSection";
import FundsTable from "./FundsTable";
import {useEffect, useState} from "react";


function PoolContent() {

    const endpoint = "";
    const apiKey = "";
    const paypalLink = "";

    const [mpData, setMpData] = useState(null);

    useEffect(() => {
       let mpName = getMpName();
       if(mpName === null) {
           return;
       }
        fetch(endpoint + mpName, {
            "headers": {
                "x-api-key": apiKey
            }
        }).then(resp => resp.json()).then(resp => {
            setMpData(resp)
        })
    }, [])

    return(
        <Container id={'content'}>
            <InfoSection data={mpData} paypalLink={paypalLink}/>
            {mpData != null && <FundsTable data={mpData}/>}
            {mpData === null &&
                <Center>
                    <Stack w={'80%'}>
                        <Skeleton height='40px' />
                        <Skeleton height='40px' />
                        <Skeleton height='40px' />
                    </Stack>
                </Center>
            }
        </Container>
    )

    function getMpName() {
        let split = window.location.href.split("?");
        if(split === null || split.length < 2) {
            return null;
        }
        let params = split[1].split("&");
        for(let i = 0; i < params.length; i++) {
            let param = params[i];
            let kv = param.split("=");
            if (kv[0].toLowerCase() === "mp") {
                return kv[1];
            }
        }
        return null;
    }
}

export default PoolContent