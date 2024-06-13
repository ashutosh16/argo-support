import React from 'react';
import './styles.scss';
import {Panel} from "./component/panel/panel";


export const Extension = (props: {
    application: any;
    openFlyout: (() => void);

}) => {

    const handleClick = () => {
        props.openFlyout();
    };

    return (
        <div className="application-status-panel__item">
            <div className="genai-container">
                <div className="genai-image-container" style={{ position: 'relative' }} onClick={handleClick}>
                    <img className="genai-image" src="https://raw.githubusercontent.com/ashutosh16/argo-support/main/ui/src/images/genai.svg"/>
                    <span style={{
                        position: 'absolute',
                        top: '50%',
                        left: '50%',
                        transform: 'translate(-50%, -50%)',
                        color: '#1a1a1a',
                        fontSize: '14px',
                        fontWeight: 'bold'
                    }}>GenAI</span>
                    <span className="genai-footer-notes">Click to start GenAI (Beta v0.1)</span> </div>  </div>

        </div>
    );

};

export const Flyout = (props: {
    application: any;
    tree: any;
}) => {
    return (
        <React.Fragment>
            <div>
                <Panel
                    application={props.application}
                    tree={props.tree}
                />
            </div>
        </React.Fragment>
    );
};

export const App = Extension;
