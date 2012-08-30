<form class='form-inline'>
<% if (data.id) { %>
	<input type=hidden name='id' value='<%= data.id %>'>
<% } %>
<input class='input-xxlarge' placeholder='name' type=text name='name' value='<%= (data.name || '') %>'>
<input class='btn btn-primary save' type='button' name='_go' value='Save!' data-type='<%= data._type %>'>
</form>
