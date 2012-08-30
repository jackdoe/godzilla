<form class='form-inline'>
<% if (data.id) { %>
	<input type=hidden name='id' value='<%= data.id %>'>
<% } %>
<input class='input-xxlarge' placeholder='title' type=text name='title' value='<%= (data.title || '') %>'><br>
<textarea class='input-xxlarge' placeholder='long text' rows='20' name='long'><%= (data.long || '') %></textarea><br>
<input class='btn btn-primary save' type='button' name='_go' value='Save!' data-type='<%= data._type %>'>
</form>