// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.819
package partials

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import "sd/pkg/types"

func ProfileAddDialog(instance types.Instance, device types.Device) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<div class=\"fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center\" id=\"modal-backdrop\" hx-target=\"this\" hx-swap=\"outerHTML\" _=\"on keyup[key==&#39;Escape&#39;] trigger click on &lt;button[hx-get=&#39;/partials/profile/close-dialog&#39;]/&gt;\n\t\t   on click if event.target.id == &#39;modal-backdrop&#39; trigger click on &lt;button[hx-get=&#39;/partials/profile/close-dialog&#39;]/&gt;\" tabindex=\"0\" autofocus><div class=\"bg-sd-dark p-6 rounded-lg shadow-xl w-96\"><h2 class=\"text-xl font-semibold mb-4 text-white\">Add New Profile</h2><form hx-post=\"/api/profile/create\" hx-target=\"#profile-card-list\" hx-swap=\"innerHTML\" class=\"space-y-4\"><input type=\"hidden\" name=\"instanceId\" value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(instance.ID)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/partials/profile_add_dialog.templ`, Line: 24, Col: 62}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "\"> <input type=\"hidden\" name=\"deviceId\" value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var3 string
		templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(device.ID)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/partials/profile_add_dialog.templ`, Line: 25, Col: 58}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "\"><div><label class=\"block text-sm font-medium text-gray-300 mb-2\">Profile Name</label> <input autofocus type=\"text\" name=\"name\" class=\"w-full p-2 bg-sd-lighter text-black rounded border border-sd-light focus:outline-none focus:border-blue-500\" placeholder=\"Enter profile name\" required></div><div class=\"flex justify-end gap-2\"><button type=\"button\" class=\"px-4 py-2 bg-sd-light text-white rounded hover:bg-sd-lighter transition-colors\" hx-get=\"/partials/profile/close-dialog\" hx-target=\"#modal-backdrop\" hx-swap=\"outerHTML\">Cancel</button> <button type=\"submit\" class=\"px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors\">Create</button></div></form></div></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
