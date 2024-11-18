import React from "react";
import Counter from "./components/Counter";

function App(props) {
    console.log("APP rendered", props);
    return (
        <div>
            <h1>title:{props.Name}</h1>
            <Counter defaultNum={props.InitialNumber}/>
        </div>
    );
}

export default App;
