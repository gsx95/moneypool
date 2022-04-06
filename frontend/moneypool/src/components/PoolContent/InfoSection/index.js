import {Center, Container, Image, Link, Skeleton, Tag, Text} from "@chakra-ui/react";
import logo from "../../../money.png";

function InfoSection(props) {

    const dataLoaded = props.data !== null;

    let transactions = (props.data === null) ? [] : props.data["transactions"];
    let name = (props.data === null) ? "" : props.data["name"];
    let title = (props.data === null) ? "" : props.data["title"];
    let open = (props.data === null) ? "" : props.data["open"];
    let paypalLink = props.paypalLink;

    let sum = 0;
    transactions.forEach(tr => {
        let base = tr["base"];
        let fraction = tr["fraction"];
        sum += base + (fraction/100);
    });

    const sumText = parseFloat(sum).toFixed(2) + "€";

    return (
        <Container>
            <Image className={'centered'} src={logo} id={'logo'}/>
            {dataLoaded && <Text id={'sum-text'} fontSize='3xl'>{sumText}</Text>}
            {!dataLoaded && <Center><Skeleton id={'sum-text-skel'} w={"100px"} h={"40px"}/></Center>}
            <Text id={'contrib-text'} fontSize='xs'>contributed for</Text>
            {dataLoaded && <Text id={'title-text'} fontSize='3xl'>{title}</Text>}
            {!dataLoaded && <Center><Skeleton id={'title-text-skel'} w={"300px"} h={"30px"}/></Center>}
            {dataLoaded && <Center id={'tags-stack'}>
                <Tag className={'tag'} variant='solid' size={'sm'} colorScheme={open ? 'teal':'red'}>{open ? 'Open' : 'Closed'}</Tag>
                <Tag className={'tag'} variant='solid' size={'sm'} colorScheme='teal'>No Goal</Tag>
                <Tag className={'tag'} variant='solid' size={'sm'} colorScheme='teal'>No Limit</Tag>
            </Center>}
            <hr id={'hr-top'}/>
            {dataLoaded && !open && <Text fontSize={'md'}>Thank you for your contributions!</Text>}
            {dataLoaded && open && <Text fontSize='md'>Want to contribute?</Text>}
            {dataLoaded && open && <Text fontSize='md'>Send funds via <Link color='teal' href={paypalLink} isExternal={true}>
                PayPal
            </Link> and start your message with '{name}'.</Text>}
            <hr id={'hr-bottom'}/>
        </Container>
    )
}

export default InfoSection;