import {Box, Button, HStack, useColorMode} from "@chakra-ui/react";
import {useColorScheme} from "react-native";
import {useState} from "react";
import { FaMoon, FaSun, FaGithub } from 'react-icons/fa';

function ControlsBox() {
    const { colorMode, toggleColorMode } = useColorMode();
    const colorScheme = useColorScheme();
    const [initedColorMode, setInitedColorMode] = useState(false);

    const colorModeInStorage = localStorage.getItem("chakra-ui-color-mode") !== null;

    if(!initedColorMode) {
        window.document.body.style = {};
        if(!colorModeInStorage && colorMode !== colorScheme) {
            toggleColorMode();
            setInitedColorMode(true);
        }
    }

    function getColorIcon(currentMode) {
        if(currentMode === "light") {
            return <FaMoon/>;
        }
        return <FaSun/>;
    }

    let smallScreen = false;
    const mq = window.matchMedia( "(min-width: 600px)" );
    smallScreen = !mq.matches;

    return (
        <div id={"controls-box"}>
            <Box>
                <HStack>
                    {smallScreen &&
                        <Button size={"sm"} onClick={toggleColorMode}>
                            {getColorIcon(colorMode)}
                        </Button>
                    }
                    {!smallScreen &&
                        <Button size={"sm"} onClick={toggleColorMode} leftIcon={getColorIcon(colorMode)}>
                            {colorMode === 'light' ? 'Dark' : 'Light'} Mode
                        </Button>
                    }
                    <Button as={"a"} size={"sm"} href={"https://github.com/gsx95/moneypool"} target="_blank" rel="noopener noreferrer"><FaGithub/></Button>
                </HStack>
            </Box>
        </div>
    )
}

export default ControlsBox;