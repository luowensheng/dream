package dream

const jsCONTENT = `

// https://blog.risingstack.com/writing-a-javascript-framework-sandboxed-code-evaluation/. Eval alternatives
const DOM_VARIBLES = new Map()

function setDOMVariable(name, value) {
    DOM_VARIBLES.set(name, value)
}

function getDOMVariable(name) {
    return DOM_VARIBLES.get(name) || null
}

function getAllDOMVariables() {
    return JSON.stringify([...DOM_VARIBLES.keys()].map(k => {
        return {name:k, value: JSON.stringify(DOM_VARIBLES.get(k)) }
    }))
}


function showVariables() {
    console.log(JSON.parse(getAllDOMVariables()))
}
async function callBackend(funcName, params, event) {
    const response = await fetch("/api/v1/call", {
        method: "POST",
        body: JSON.stringify({"call": funcName, "params": params || {}})
    });
    const data = await response.json();

    if (data.success){
        handleServerCommand(data.commands, event)
    } else {
        console.error(data.message)
    }
}

function handleServerCommand(commands, event){
    if (!commands){
        return;
    }
    console.log({commands})
    let onComplete = [];

    for (let command of commands ){
        let key = command.type;
        console.log({command})
        try {
            switch (key) {
                case "execute":
                    let result = eval(command.command);
                    console.log({result});
                    break;
                case "execute-after":
                    console.log({after: command.command});
                    onComplete.push(()=>eval(command.command));
                    break;
                case "html":
                    resetHTML(command.command);
                    break;                          
                default:
                    console.error({key, command: command.command})
                    break;
            } 
        } catch (error) {
            console.error({error})
            console.log({command})
            callBackend('error', {error: error.toString(), command})
        }

    }

    for (const f of onComplete) {
        f()
    }
}

`