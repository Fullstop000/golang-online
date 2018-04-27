import React, { Component } from 'react';
import CodeMirror from 'codemirror';
import {
  default as AnsiUp
} from 'ansi_up';
import './App.css';

const ansi_up = new AnsiUp();
let logZone;
let codeMirror;
class App extends Component {
  constructor(prop) {
    super(prop);
    this.handleRun = this.handleRun.bind(this);
    this.handleClear = this.handleClear.bind(this);
  }
  handleRun() {
    this.handleClear()
    const ws = new WebSocket("ws://localhost:8080/ws/go", "test_protocol")
    ws.onopen = () => {
      ws.send(codeMirror.getValue())
    }
    ws.onmessage = (event) => {
      let logEle = document.createElement('div')
      logEle.setAttribute('class','log-item')
      logEle.innerHTML = ansi_up.ansi_to_html(event.data)
      logZone.appendChild(logEle)
    }
    ws.onclose = (event) => {
      console.log(event.reason)
    }
  }
  handleClear() {
    logZone.innerHTML= ""
  }
  componentDidMount() {
    codeMirror = CodeMirror(document.querySelector('.code'), {
      value: `package main

import (
  "fmt"
)

func main() {
    fmt.Println("Hello World!")
}
      `,
      lineNumbers: true,
      mode: 'text/x-go',
    });
    logZone = document.querySelector('.log')
  }
  render() {
    return (
      <div className="App">
        <div className="code">
        </div>
        <div className="toolbar">
          <div className="run" onClick={this.handleRun}>RUN</div>
          <div className="clear" onClick={this.handleClear}>CLEAR</div>
        </div>
        <div className="log">

        </div>
      </div>
    );
  }
}

export default App;
