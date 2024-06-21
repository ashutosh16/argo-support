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
        <div className="application-status-panel__item" style={{display:"flex", flexDirection: "column", padding: '20px'}}>
            <div className="genai-container">
                <div className="genai-image-container" >
                    <button style={{position: 'relative', minWidth: '120px', minHeight: '20px' }}  className={`argo-button argo-button--base`}  onClick={handleClick}>
                    <svg style={{display: "flex", alignSelf: "flex-start"}} width="20" height="20" viewBox="0 0 14 14" fill="none" color="currentColor" focusable="false" aria-hidden="false"
                         xmlns="http://www.w3.org/2000/svg">
                        <path fillRule="evenodd" clipRule="evenodd"
                          d="M6.47659 5.73635C6.30749 5.73628 6.13651 5.70019 5.97403 5.62808L2.70243 4.16458L3.16386 3.1237L6.43546 4.5872C6.46855 4.6024 6.5007 4.5986 6.53001 4.5777C6.55932 4.55586 6.57256 4.52737 6.56878 4.49128L6.19339 0.911819L7.32143 0.793106L7.69587 4.37067C7.7422 4.81418 7.55499 5.23585 7.19567 5.49892C6.98022 5.65648 6.72983 5.73625 6.47659 5.73635ZM6.47659 5.73635C6.47643 5.73635 6.47627 5.73635 6.47611 5.73635H6.47706C6.4769 5.73635 6.47674 5.73635 6.47659 5.73635ZM8.7745 6.93638C8.98876 7.09489 9.23864 7.17772 9.49204 7.17986L9.49206 7.17796C9.6613 7.17938 9.83274 7.14569 9.99598 7.07393L13.2798 5.63803L12.8271 4.5933L9.54334 6.0292C9.51012 6.04412 9.478 6.04004 9.44886 6.01891C9.41973 5.99777 9.40674 5.96821 9.41083 5.93216L9.81538 2.35788L8.68838 2.22967L8.28383 5.80395C8.23375 6.24801 8.41742 6.67124 8.7745 6.93638ZM10.5202 12.7369L8.14791 10.0424C7.85395 9.70836 7.7615 9.25546 7.9033 8.83218C8.04415 8.40889 8.38821 8.10326 8.82396 8.01404L12.3337 7.29666L12.5593 8.41365L9.04954 9.13104C9.01355 9.13832 8.99067 9.15901 8.97901 9.19404C8.96735 9.22908 8.97371 9.25953 8.99804 9.28729L11.3703 11.9818L10.5202 12.7369ZM3.88862 12.927L2.90835 12.3534L4.71093 9.24515C4.93459 8.85946 5.33353 8.63018 5.77793 8.63185C6.22234 8.63352 6.61954 8.8658 6.84029 9.25316L8.61942 12.3749L7.63486 12.9411L5.85572 9.81929C5.82003 9.75648 5.72832 9.75518 5.69215 9.81867L3.88957 12.927L3.88862 12.927ZM1.83936 9.65602L2.67461 10.4276L2.67459 10.4295L5.09895 7.78171C5.39937 7.45249 5.49967 7.00242 5.3671 6.57646C5.23452 6.1505 4.89648 5.83823 4.46255 5.74053L0.967464 4.95489L0.720161 6.06728L4.21527 6.85102C4.25017 6.85899 4.27357 6.88109 4.28456 6.91539C4.29553 6.95065 4.28858 6.98192 4.26371 7.00825L1.83936 9.65602Z" fill="currentColor">
                    </path>
                    </svg>
                        <span style={{
                            position: 'relative',
                            color: '#e8e8e8',
                            fontSize: '12px',
                            fontWeight: 500,
                        }}>Intuit Assist</span>
                    </button>

                    <span className="genai-footer-notes">Click to start GenAI (Beta v0.1)</span>
                </div>
            </div>

        </div>
    );

};

export const Flyout = (props: {
    application: any;
    tree: any;
}) => {
    return (
        <React.Fragment>
            <div className="panel">
                <Panel
                    application={props.application}
                    tree={props.tree}
                />
            </div>

        </React.Fragment>
    );
};

export const App = Extension;
