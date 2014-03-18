/** @jsx React.DOM */

var _ = require('lodash/dist/lodash.underscore');
var React = require('react/addons');
var capture = require('./capture');
var irlmoji = require('./irlmoji');
var mixins = require('./mixins');
var emoji = require('../emoji');

var Timeline = React.createClass({

  mixins: [mixins.ScreenDimensionsMixin],

  getInitialState: function() {
    return {imagePath: null};
  },

  handleClick: function(ev) {
    this.props.app.router.go('/');
    return false;
  },

  handlePostImageUpload: function(path) {
    this.setState({imagePath: path});
  },

  handleCreateIRLMojiResponse: function(error, data) {
    if (error) {
      alert(error);
      return;
    }
    this.props.app.router.reload();
  },

  handleEmojiChoice: function(emojiKey) {
    var picture = this.state.imagePath;
    this.props.app.api.createIRLMoji(emojiKey, picture,
      this.handleCreateIRLMojiResponse);
    return false;
  },

  handleEmojiCancel: function(ev) {
    this.setState({imagePath: null});
    return false;
  },

  handleEmojiTap: function(kind, im, ev) {
    if (kind === 'user') {
      this.props.app.router.go('/user/' + im.user.username);
    } else if (kind === 'picture') {
      this.props.app.router.go('/timeline/emoji/' + emoji.getDisplay(im.emoji));
    }
    return false;
  },

  render: function() {
    var Capture = capture.Capture;
    var EmojiPicker = capture.EmojiPicker;
    var IRLMoji = irlmoji.IRLMoji;
    return (
      <div>
        <p>You are logged in! ({this.props.user.username})</p>
        <p><a href="/" onClick={this.handleClick}>Home</a></p>
        <p><a href="/logout">Logout</a></p>
        {_.map(this.props.timeline.timeline, function(im) {
          return (
            <IRLMoji key={im.id}
                     irlmoji={im}
                     app={this.props.app}
                     screenWidth={this.state.screenWidth}
                     screenHeight={this.state.screenHeight}
                     onEmojiTap={this.handleEmojiTap} />
          );
        }, this)}
        {this.state.imagePath ?
          <EmojiPicker app={this.props.app}
                       onCancel={this.handleEmojiCancel}
                       onChoice={this.handleEmojiChoice} /> :
          <Capture app={this.props.app}
                   onPostImageUpload={this.handlePostImageUpload} />}
      </div>
    );
  }

});

module.exports = {
  Timeline: Timeline
};