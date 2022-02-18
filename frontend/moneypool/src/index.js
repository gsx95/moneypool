import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';
import {ChakraProvider, ColorModeScript} from "@chakra-ui/react";
import {ThemeProvider} from "@emotion/react";
import theme from "./theme";

ReactDOM.render(
    <ChakraProvider >
        <ThemeProvider theme={theme}>
            <ColorModeScript initialColorMode={theme.config.initialColorMode} />
            <App />
        </ThemeProvider>
    </ChakraProvider>,
  document.getElementById('root')
);