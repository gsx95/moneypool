import './app.css'

import PoolContent from "./components/PoolContent";
import ControlsBox from "./components/ControlsBox";

function App() {

  return (
      <div id={"main"}>
          <ControlsBox/>
          <PoolContent/>
      </div>
  );
}

export default App;
