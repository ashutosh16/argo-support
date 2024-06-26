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
                            <svg xmlns="http://www.w3.org/2000/svg" width="20px" height="20px" fill="none" viewBox="0 0 24 24" color="currentColor" focusable="false" aria-hidden="true" className=""><path fill="currentColor" d="M19 9h-3.69l.28-1.62a4 4 0 0 0-2.14-4.27 1 1 0 0 0-1.22.25L7.53 9H3a1 1 0 0 0-1 1v10a1 1 0 0 0 1 1h13.68a3 3 0 0 0 2.76-1.82l2.28-5.32a3.39 3.39 0 0 0 .28-1.37V12a3 3 0 0 0-3-3ZM7 19H4v-8h3v8Zm13-6.51c0 .203-.041.403-.12.59l-2.28 5.31a1 1 0 0 1-.92.61H9v-8.64l4.16-5A2 2 0 0 1 13.63 7l-.5 2.8a.999.999 0 0 0 1 1.17H19a1 1 0 0 1 1 1v.52Z"></path></svg>
                        </a>
                        <a className={`downvote ${vote === 'down' ? 'selected' : ''}`} onClick={handleTumbsDownClick}>
                            <svg xmlns="http://www.w3.org/2000/svg" width="20px" height="20px" fill="none" viewBox="0 0 24 24" color="currentColor" focusable="false" aria-hidden="true" className=""><path fill="currentColor" d="M19.44 4.82A3 3 0 0 0 16.68 3H3a1 1 0 0 0-1 1v10a1 1 0 0 0 1 1h4.53l4.7 5.64a1 1 0 0 0 1.22.25 4 4 0 0 0 2.14-4.27L15.31 15H19a3 3 0 0 0 3-3v-.49a3.39 3.39 0 0 0-.28-1.37l-2.28-5.32ZM7 13H4V5h3v8Zm13-1a1 1 0 0 1-1 1h-4.88a1 1 0 0 0-1 1.17l.5 2.8a2 2 0 0 1-.47 1.66L9 13.64V5h7.68a1 1 0 0 1 .92.61l2.28 5.31c.079.187.12.387.12.59V12Z"></path></svg>
                        </a>
                    </div>
                    {showResponseBox && (
                        <div className="modal-container">
                            <div className="modal-content">
                                <span style={{paddingLeft: '50px'}}>
                                    <i className="fas fa-times" onClick={() => setShowResponseBox(false)}></i>
                                    </span>
                            <div className="modal-header">
                                <span className="modal-header-icon">
                                {vote === 'up' ? (
                                        <svg xmlns="http://www.w3.org/2000/svg" width="20px" height="20px" fill="none" viewBox="0 0 24 24" color="currentColor" focusable="false" aria-hidden="true" className=""><path fill="currentColor" d="M19 9h-3.69l.28-1.62a4 4 0 0 0-2.14-4.27 1 1 0 0 0-1.22.25L7.53 9H3a1 1 0 0 0-1 1v10a1 1 0 0 0 1 1h13.68a3 3 0 0 0 2.76-1.82l2.28-5.32a3.39 3.39 0 0 0 .28-1.37V12a3 3 0 0 0-3-3ZM7 19H4v-8h3v8Zm13-6.51c0 .203-.041.403-.12.59l-2.28 5.31a1 1 0 0 1-.92.61H9v-8.64l4.16-5A2 2 0 0 1 13.63 7l-.5 2.8a.999.999 0 0 0 1 1.17H19a1 1 0 0 1 1 1v.52Z"></path></svg>
                                    ) :
                                    (
                                        <svg xmlns="http://www.w3.org/2000/svg" width="20px" height="20px" fill="none" viewBox="0 0 24 24" color="currentColor" focusable="false" aria-hidden="true" className=""><path fill="currentColor" d="M19.44 4.82A3 3 0 0 0 16.68 3H3a1 1 0 0 0-1 1v10a1 1 0 0 0 1 1h4.53l4.7 5.64a1 1 0 0 0 1.22.25 4 4 0 0 0 2.14-4.27L15.31 15H19a3 3 0 0 0 3-3v-.49a3.39 3.39 0 0 0-.28-1.37l-2.28-5.32ZM7 13H4V5h3v8Zm13-1a1 1 0 0 1-1 1h-4.88a1 1 0 0 0-1 1.17l.5 2.8a2 2 0 0 1-.47 1.66L9 13.64V5h7.68a1 1 0 0 1 .92.61l2.28 5.31c.079.187.12.387.12.59V12Z"></path></svg>
                                    )
                                }
                                </span>
                                <h5 className="feedback-header">Thanks for the feedback!</h5>
                                Why did you choose this answer? (Optional)
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