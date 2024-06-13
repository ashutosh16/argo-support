import React, { useState, useEffect } from 'react';
import './feedback.scss';
import * as Const from '../../shared/constants';

export const FeedbackComponent = (props) => {
    const [vote, setVote] = useState(null);
    const [text, setText] = useState('');
    const [username, setUsername] = useState(null);
    const [showResponseBox, setShowResponseBox] = useState(false);

    useEffect(() => {
        fetchUserInfo();
    }, []);

    const   fetchUserInfo = async () => {

            const response = await fetch(Const.APIs.fetchUserInfo());
            if (response.ok) {
                const data = await response.json()
                setUsername(response?data?.username: {});
            } else {
                console.log('user info not exist');
            }

    };

    const patchAnnotation = async () => {
        const url = Const.APIs.patchAnnotation(props.applicationName, props.applicationNamespace, props.destNamespace);
        const patch = JSON.stringify({
            "metadata": {
                "annotations": {
                    "argosupport.argoproj.extensions.io/genai": `${JSON.stringify(props.application.status)}`
                }
            }
        });
        try {
            const response = await fetch(url, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(patch)
            });

            if (!response.ok) throw new Error('error patching the genai resource');

            const patchedData = await response.json();
            await new Promise(resolve => setTimeout(resolve, 800));
            console.log('Patch request successful:', patchedData);
        } catch (error) {
            console.error('Error in patch request:', error);
        }
    };

    const handleTumbsUpClick = () => {
        setVote('up');
        setShowResponseBox(true);
    };

    const handleTumbsDownClick = () => {
        setVote('down');
        setShowResponseBox(true);
    };

    const handleTextChange = (event) => {
        setText(event.target.value);
    };

    const handleSubmit = () => {
         patchAnnotation()
        setShowResponseBox(false);
    };

    return (
        <div className="feedback">
            Was this helpful? &nbsp;
            <div className={`votes ${showResponseBox ? 'disabled' : 'enabled'}`}>
                <a className={`upvote ${vote === 'up' ? 'selected' : ''}`} onClick={handleTumbsUpClick} style={{ marginRight: '10px' }}>
                    <i style={{ fontSize: "14px" }} className="fas fa-thumbs-up"></i>
                </a>
                <a className={`downvote ${vote === 'down' ? 'selected' : ''}`} onClick={handleTumbsDownClick}>
                    <i style={{ fontSize: "14px" }} className="fas fa-thumbs-down"></i>
                </a>
            </div>
            {showResponseBox && (
                <div className="modal-content">
                    {vote === 'up' ? <i className="fas fa-thumbs-up" style={{color: "#ffe24f", fontSize: "24px"}}></i> : <i  style={{color: "#ffe24f", fontSize: "24px"}} className="fas fa-thumbs-down"></i>}
                    <div className="modal-header">
                        Would you like to tell us more?
                        <i className="fas fa-times" onClick={() => setShowResponseBox(false)}></i>
                    </div>
                    <div className="modal-body">
                        <textarea id="modalCommentArea" className="modal-comment-area" value={text} onChange={handleTextChange} placeholder="Enter your feedback here..." />
                        <button className="modal-submit-button" onClick={handleSubmit}>
                            <span>Send</span>
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};
