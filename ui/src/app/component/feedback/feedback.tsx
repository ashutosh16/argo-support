import React, { useState, useEffect } from 'react';
import './feedback.scss';
import * as Const from '../../shared/services/constants';

export const    FeedbackComponent = (props: FeedBackProps) => {
    const [vote, setVote] = useState(null);
    const [msg, setMsg] = useState('');
    const [username, setUsername] = useState(null);
    const [showResponseBox, setShowResponseBox] = useState(false);

    useEffect(() => {
        fetchUserInfo();
    }, []);

    const fetchUserInfo = async () => {
        const response = await fetch(Const.APIs.getUserInfo());
        if (response.ok) {
            const data = await response.json();
            setUsername(data?.username || '');
        } else {
            console.log('user info not exist');
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
        setMsg(event.target.value);
    };

    const handleSubmit = () =>  {
        const patch = JSON.stringify({
            "metadata": {
                "annotations": {
                    "argosupport.argoproj.extensions.io/wf-feedback": JSON.stringify({
                        "name": props.result.name,
                        "user": username,
                        "vote": vote,
                        "message": msg
                    })
                },
            }
        });
        props.patchApi(patch);
        setShowResponseBox(false);
    };

    const feedback = props.result?.feedback || {};

    return (
        <div className="feedback">
            {feedback ? (
                <>
                    <div className={`votes ${showResponseBox ? 'disabled' : 'enabled'}`}>
                        Is this analysis helpful? &nbsp;
                        <a className={`upvote ${vote === 'up' ? 'selected' : ''}`} onClick={handleTumbsUpClick}>
                            <i style={{ fontSize: "14px" }} className="fas fa-thumbs-up"></i>
                        </a>
                        <a className={`downvote ${vote === 'down' ? 'selected' : ''}`} onClick={handleTumbsDownClick}>
                            <i style={{ fontSize: "14px" }} className="fas fa-thumbs-down"></i>
                        </a>
                    </div>
                    {showResponseBox && (
                        <div className="modal-content">
                            {vote === 'up' ? (
                                <i className="fas fa-thumbs-up" style={{ color: "#ffe24f", fontSize: "24px" }}></i>
                            ) : (
                                <i className="fas fa-thumbs-down" style={{ color: "#ffe24f", fontSize: "24px" }}></i>
                            )}
                            <div className="modal-header">
                                Tell me more!
                                <i className="fas fa-times" onClick={() => setShowResponseBox(false)}></i>
                            </div>
                            <div className="modal-body">
                                <textarea
                                    id="modalCommentArea"
                                    className="modal-comment-area"
                                    value={msg}
                                    onChange={handleTextChange}
                                    placeholder="Enter your feedback here..."
                                />
                                <button className="modal-submit-button" onClick={handleSubmit}>
                                    <span>Send</span>
                                </button>
                            </div>
                        </div>
                    )}
                </>
            ) : (
                'Thank you for your feedback!'
            )}

        </div>
    );
};

interface FeedBackProps {
    result: any;
    patchApi: ( patch: string) => any;
}